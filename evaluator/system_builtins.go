package evaluator

import (
	"archive/zip"
	"base/object"
	"io"
	"os"
	"path/filepath"

	"github.com/robfig/cron/v3"
)

var globalCron = cron.New()

func RegisterSystemBuiltins() {
	globalCron.Start()

	builtins["schedule"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			spec, ok1 := args[0].(*object.String)
			fn, ok2 := args[1].(*object.Function)

			if !ok1 || !ok2 {
				return newError("arguments to `schedule` must be (STRING, FUNCTION)")
			}

			_, err := globalCron.AddFunc(spec.Value, func() {
				applyFunction(env, fn, []object.Object{})
			})

			if err != nil {
				return newError("cron schedule error: %s", err.Error())
			}

			return TRUE
		},
	}

	builtins["archive.zip"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			source, ok1 := args[0].(*object.String)
			target, ok2 := args[1].(*object.String)

			if !ok1 || !ok2 {
				return newError("arguments to `archive.zip` must be (STRING, STRING)")
			}

			err := zipSource(source.Value, target.Value)
			if err != nil {
				return newError("archive error: %s", err.Error())
			}

			return TRUE
		},
	}
}

func zipSource(source, target string) error {
	zipFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name, err = filepath.Rel(filepath.Dir(source), path)
		if err != nil {
			return err
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}
