package utils

import (
	"io"
	"log"
	"os"
)

// logが実行されたときにファイル上にこの関数の引数においたファイル名が作成されてそこにログが追加されていく
func LoggingSettings(logFile string){
	// ログファイルの設定
	logfile, err := os.OpenFile(logFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
		// O_RDWR   // open the file read-write.
		// O_APPEND // append data to the file when writing.
		// O_CREATE // create a new file if none exists.
	if err != nil{
		log.Fatalf("file=logFile err=%s", err.Error())
	}
	// stdout(画面表示用), logfile(ファイルへの記録用)の両方に書き込むのでmultiwriteを使う
	multiLogFile := io.MultiWriter(os.Stdout, logfile)
	// ログファイルの日付やファイル内容のカスタマイズ→ターミナル上での表示の仕方などを変える
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile) 
	log.SetOutput(multiLogFile)
}