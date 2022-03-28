package manager

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

const (
	ruby = "19b4554491cd"
	node_js = "224197962b1b"
	c = "94e07ec1ec23"
)

type Code struct {
	TaskID       string `json:"task_id"`
	Code         string `json:"code"`
	CompilerType string `json:"compiler_type"`
}

func Start(code Code) string {
	enc, _ := base64.StdEncoding.DecodeString(code.Code)
	compilerType, filetype := CheckCompilerType(code.CompilerType)
	ctx := context.Background()

	// 1 taskあたり10sec * 3  +1 = 31秒を最大コンテナ生存時間にする
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	MaxMemorySize := "512M" // メモリ制限 コンテナ1つ512メガバイトまで
	mem, _ := strconv.ParseInt(MaxMemorySize, 10, 64)

	// コンテナを作る
	f := false
	swappness := int64(0)   // スワップを封印
	PidsLimit := int64(512) // 対フォーク爆弾の設定

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:           compilerType, // Done: Taskに合わせてイメージを変える
		NetworkDisabled: true,
		//Cmd:   []string{"tail", "-f", "/dev/null"},
		Cmd: []string{"/main", code.CompilerType, code.TaskID}, // 実行する時のコマンド
		Tty: false,                                             // Falseにしておく
	}, &container.HostConfig{
		AutoRemove:  false, // これをOnにするとLogが取れなくなって死ぬ
		NetworkMode: "none",
		Resources: container.Resources{
			Memory:           mem,
			MemorySwap:       mem,
			OomKillDisable:   &f,
			MemorySwappiness: &swappness,
			PidsLimit:        &PidsLimit,
		},
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	// 送られたファイルをコンパイラタイプに合わせてtarにまとめる
	// ToDo: 名前がかぶらないようにする
	createCodeTarfile(enc, filetype)

	// ToDo: テストケースも実行時にコピーするようにする
	archive, _ := os.Open("docker.tar")
	defer archive.Close()
	err = cli.CopyToContainer(ctx, resp.ID, "/", bufio.NewReader(archive), types.CopyToContainerOptions{})
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return "Error"
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	_ = cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		RemoveLinks:   false,
		Force:         true,
	})

	r, e := ioutil.ReadAll(out)
	if e != nil {
		panic(e)
	}
	fmt.Println(out)
	if len(string(r)) < 8 {
		fmt.Println(r)
		return "Err"
	}
	return string(r)[8:]
}

func createCodeTarfile(file []byte, filetype string) {
	err := ioutil.WriteFile("main"+filetype, file, 0666)
	if err != nil {
		panic(err)
	}

	tarfile, _ := os.Create("docker.tar")
	code, err := os.Open("main" + filetype)
	defer code.Close()
	tarWriter := tar.NewWriter(tarfile)
	defer tarWriter.Close()

	c, _ := code.Stat()
	if err := tarWriter.WriteHeader(&tar.Header{
		Name:    "main" + filetype,
		Mode:    int64(c.Mode()),
		ModTime: c.ModTime(),
		Size:    c.Size(),
	}); err != nil {
		panic(err)
	}

	f, _ := os.Open("main" + filetype)
	defer f.Close()

	if _, err := io.Copy(tarWriter, f); err != nil {
		panic(err)
	}
}


func CheckCompilerType(CompilerType string) (string, string) {
	switch CompilerType {
	case "c-gcc":
		return c, ".c"
	case "c-clang":
		return c, ".c"
	case "cxx-gxx":
		return c, ".cpp"
	case "cxx-clang":
		return c, ".cpp"
	case "node-js":
		return node_js, ".js"
	case "ruby":
		return ruby, ".rb"
	default:
		break
	}
	return "", ""
}
