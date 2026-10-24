package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yumeng-tech/ai-collab-board/backend/internal/botclient"
)

// viewFlags 挂到子命令 FlagSet：--full / --id / --json
func viewFlags(fs *flag.FlagSet) (full, ids, asJSON *bool) {
	full = fs.Bool("full", false, "接近 REST 全量 JSON（含裸 UUID）")
	ids = fs.Bool("id", false, "仍瘦，但附加裸 id（比 --full 轻）")
	asJSON = fs.Bool("json", false, "强制瘦 JSON（非文本表）")
	return
}

func optsFrom(full, ids, asJSON bool) botclient.ViewOpts {
	return botclient.ViewOpts{Full: full, IDs: ids, JSON: asJSON}
}

func optsDetail(full, ids, asJSON bool) botclient.ViewOpts {
	o := optsFrom(full, ids, asJSON)
	o.Detail = true
	return o
}

func printView(opts botclient.ViewOpts, text string, data any) {
	if opts.Full || opts.JSON {
		if opts.Full {
			// data should already be full-shaped
			fmt.Print(botclient.MarshalView(data))
			return
		}
		fmt.Print(botclient.MarshalView(data))
		return
	}
	fmt.Print(text)
}

func printRawFull(b json.RawMessage) {
	var v any
	if json.Unmarshal(b, &v) == nil {
		fmt.Print(botclient.MarshalView(v))
		return
	}
	fmt.Println(string(b))
}

// discardFS 静默 FlagSet（错误由调用方处理）。
func newFS(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

// reorderFlags 把 -/-- 旗标挪到位置参数之前（Go flag 在首个非旗标处停解析）。
func reorderFlags(args []string) []string {
	var flags, pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			pos = append(pos, args[i+1:]...)
			break
		}
		if len(a) > 1 && a[0] == '-' {
			flags = append(flags, a)
			// boolean flags have no value; value flags look like --k=v or --k v
			if !containsEq(a) && i+1 < len(args) && !looksFlag(args[i+1]) && flagTakesValue(a) {
				i++
				flags = append(flags, args[i])
			}
			continue
		}
		pos = append(pos, a)
	}
	return append(flags, pos...)
}

func looksFlag(s string) bool {
	return len(s) > 1 && s[0] == '-'
}

func containsEq(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return true
		}
	}
	return false
}

// flagTakesValue：已知带值旗标（非 --full/--id/--json/--leaf-only 等 bool）。
func flagTakesValue(name string) bool {
	for len(name) > 0 && name[0] == '-' {
		name = name[1:]
	}
	switch name {
	case "full", "id", "json", "leaf-only", "unassigned", "autosave", "h", "help":
		return false
	default:
		return true
	}
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
