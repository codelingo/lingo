package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/codegangsta/cli"
	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
	"github.com/mholt/archiver"
)

func init() {
	register(&cli.Command{
		Name:   "install",
		Usage:  "Install an Action",
		Action: installAction,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "owner",
				Value: "codelingo",
				Usage: "Owner of the Flow",
			},
		},
	}, false, false)
}

func installAction(ctx *cli.Context) {
	if err := install(ctx); err != nil {
		util.FatalOSErr(err)
		return
	}
	fmt.Printf("Success! You can now run the installed flow with `lingo run %s`\n", ctx.Args()[0])
}

func install(c *cli.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return errors.New("Failed to install Flow - no Flow given.")
	}

	home, err := util.LingoHome()
	if err != nil {
		return errors.Trace(err)
	}

	ownerName := c.String("owner")
	flowName := args[0]
	flowPath := fmt.Sprintf("%s/flows/%s/%s", home, ownerName, flowName)

	if err := os.MkdirAll(flowPath, 0755); err != nil {
		return errors.Trace(err)
	}

	// TODO(waigani) pass in version
	version := "0.0.0"

	currentOS := runtime.GOOS
	archiveName := "cmd.tar.gz"
	if currentOS == "windows" {
		archiveName = "cmd.exe.zip"
	}

	fileUrl := fmt.Sprintf("https://github.com/codelingo/codelingo/raw/master/flows/%s/%s/bin/%s/%s/%s/%s", ownerName, flowName, currentOS, runtime.GOARCH, version, archiveName)
	fmt.Println("Installing Flow:", fileUrl)
	if err := DownloadFile(flowPath+"/"+archiveName, fileUrl); err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(extractCMD(flowPath, archiveName))
}

func DownloadFile(filepath string, url string) error {

	out, err := os.Create(filepath)
	if err != nil {
		return errors.Trace(err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return errors.Trace(err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

func extractCMD(dir, archiveName string) error {

	var ext string
	if runtime.GOOS == "windows" {

		ext = ".exe"
		err := archiver.Zip.Open(dir+"/"+archiveName, ".")
		if err != nil {
			return errors.Trace(err)
		}

	} else {

		err := archiver.TarGz.Open(dir+"/"+archiveName, dir)
		if err != nil {
			return errors.Trace(err)
		}

	}

	return errors.Trace(os.Chmod(dir+"/cmd"+ext, 0755))
}
