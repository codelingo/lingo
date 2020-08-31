package commands

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/codelingo/lingo/app/util"
	"github.com/juju/errors"
	"github.com/mholt/archiver"
	"github.com/urfave/cli"
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
				Usage: "Owner of the Action",
			},
		},
	}, false, false)
}

func installAction(ctx *cli.Context) {
	if err := install(ctx); err != nil {
		util.FatalOSErr(err)
		return
	}
	fmt.Printf("Success! You can now run the installed Action with `lingo run %s`\n", ctx.Args()[0])
}

func install(c *cli.Context) error {
	args := c.Args()
	if len(args) == 0 {
		return errors.New("Failed to install Action - no Action given.")
	}

	home, err := util.LingoHome()
	if err != nil {
		return errors.Trace(err)
	}

	ownerName := c.String("owner")
	actionName := args[0]
	actionPath := fmt.Sprintf("%s/flows/%s/%s", home, ownerName, actionName)

	if err := os.MkdirAll(actionPath, 0755); err != nil {
		return errors.Trace(err)
	}

	// TODO(waigani) pass in version
	version := "0.0.0"

	currentOS := runtime.GOOS
	archiveName := "cmd.tar.gz"
	if currentOS == "windows" {
		archiveName = "cmd.exe.zip"
	}

	fileUrl := fmt.Sprintf("https://github.com/codelingo/actions/raw/master/actions/%s/%s/bin/%s/%s/%s/%s", ownerName, actionName, currentOS, runtime.GOARCH, version, archiveName)
	fmt.Println("Installing Action:", fileUrl)
	if err := DownloadFile(actionPath+"/"+archiveName, fileUrl); err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(extractCMD(actionPath, archiveName))
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
		err := archiver.DefaultZip.Unarchive(dir+"/"+archiveName, dir)
		if err != nil {
			return errors.Trace(err)
		}

	} else {

		err := archiver.DefaultTarGz.Unarchive(dir+"/"+archiveName, dir)
		if err != nil {
			return errors.Trace(err)
		}

	}

	return errors.Trace(os.Chmod(dir+"/cmd"+ext, 0755))
}
