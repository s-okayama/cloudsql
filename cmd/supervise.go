package cmd

import (
	"log"
	"os/exec"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// connect が内部的に使う監視コマンド。ps に cloud-sql-proxy の文字列を
// 出さないよう、-- 以降には proxy の引数だけを受け取る。
var superviseCmd = &cobra.Command{
	Use:    "supervise --ttl <duration> -- <proxy args>",
	Short:  "internal: run cloud-sql-proxy and kill it after --ttl",
	Hidden: true,
	Args:   cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ttl, err := cmd.Flags().GetDuration("ttl")
		if err != nil {
			log.Fatalf("Error getting ttl: %v", err)
		}
		superviseProcess(ttl, append([]string{"cloud-sql-proxy"}, args...))
	},
}

// argv を子プロセスとして起動し、ttl 経過でプロセスグループごと停止する。
// 子が先に終了した場合（disconnect された場合など）は即座に戻る。
func superviseProcess(ttl time.Duration, argv []string) {
	child := exec.Command(argv[0], argv[1:]...)
	child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := child.Start(); err != nil {
		log.Fatalf("supervise: failed to start %s: %v", argv[0], err)
	}

	done := make(chan struct{})
	go func() {
		_ = child.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(ttl):
		pgid := child.Process.Pid
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
		select {
		case <-done:
		case <-time.After(10 * time.Second):
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
			<-done
		}
	}
}
