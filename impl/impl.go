package impl

import (
	"better-docker-ps/cli"
	"better-docker-ps/docker"
	"better-docker-ps/printer"
	"better-docker-ps/pserr"
	"encoding/json"
	"git.blackforestbytes.com/BlackForestBytes/goext/langext"
	"git.blackforestbytes.com/BlackForestBytes/goext/mathext"
	"git.blackforestbytes.com/BlackForestBytes/goext/syncext"
	"golang.org/x/term"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func Execute(ctx *cli.PSContext) error {
	return executeSingle(ctx, false)
}

func Watch(ctx *cli.PSContext, d time.Duration) error {

	sigTermChannel := make(chan os.Signal, 8)
	signal.Notify(sigTermChannel, os.Interrupt, syscall.SIGTERM)

	for {

		err := executeSingle(ctx, true)
		if err != nil {
			return err
		}

		_, isSig := syncext.ReadChannelWithTimeout(sigTermChannel, d)
		if isSig {
			ctx.PrintPrimaryOutput("")
			ctx.PrintPrimaryOutput("Watch canceled with Ctrl+C")
			return nil
		}

	}
}

func executeSingle(ctx *cli.PSContext, clear bool) error {
	for _, fmt := range ctx.Opt.Format {
		if strings.Contains(fmt, "{{.Size}}") {
			ctx.Opt.WithSize = true
		}
	}

	jsonraw, err := docker.ListContainer(ctx)
	if err != nil {
		return err
	}

	ctx.PrintVerboseKV("API response", langext.TryPrettyPrintJson(string(jsonraw)))

	var data []docker.ContainerSchema
	err = json.Unmarshal(jsonraw, &data)
	if err != nil {
		return pserr.DirectOutput.Wrap(err, "Failed to decode Docker API response")
	}

	if len(ctx.Opt.SortColumns) > 0 {
		data = doSort(ctx, data, ctx.Opt.SortColumns, ctx.Opt.SortDirection)
	}

	for i, v := range ctx.Opt.Format {

		if clear {
			ctx.ClearTerminal()
		}

		ok, err := doOutput(ctx, data, v, i == len(ctx.Opt.Format)-1)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

	}

	return pserr.DirectOutput.New("Missing format specification for output")
}

func doSort(ctx *cli.PSContext, data []docker.ContainerSchema, skeys []string, sdirs []cli.SortDirection) []docker.ContainerSchema {

	langext.SortSliceStable(data, func(v1, v2 docker.ContainerSchema) bool {

		// return true if v1 < v2

		for i := 0; i < len(skeys); i++ {

			sfn, ok := getSortFun(skeys[i])
			if !ok {
				continue
			}

			cmp := sfn(ctx, &v1, &v2)
			if sdirs[i] == "DESC" {
				cmp = cmp * -1
			}

			if cmp < 0 {
				return true
			} else if cmp > 0 {
				return false
			}
		}

		return false // equals
	})

	return data
}

func doOutput(ctx *cli.PSContext, data []docker.ContainerSchema, format string, force bool) (bool, error) {
	if format == "idlist" {

		for _, v := range data {
			if ctx.Opt.Truncate {
				ctx.PrintPrimaryOutput(v.ID[0:12])
			} else {
				ctx.PrintPrimaryOutput(v.ID)
			}
		}
		return true, nil

	} else if strings.HasPrefix(format, "table ") {

		columns := parseTableDef(format)
		outWidth := printer.Width(ctx, data, columns)

		if !force {
			termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil && 0 < termWidth && termWidth < outWidth {
				return false, nil
			}
		}

		printer.Print(ctx, data, columns)
		return true, nil

	} else {

		lines := make([]string, 0)
		outWidth := 0

		for _, v := range data {
			str := replaceSingleLineColumnData(ctx, data, v, format)
			lines = append(lines, str)
			outWidth = mathext.Max(outWidth, printer.RealStrLen(str))
		}

		if !force {
			termWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
			if err == nil && 0 < termWidth && termWidth < outWidth {
				return false, nil
			}
		}

		for _, v := range lines {
			ctx.PrintPrimaryOutput(v)
		}
		return true, nil

	}
}
