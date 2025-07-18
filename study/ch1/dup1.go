package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"
)

func main() {
	//标准输入中多次出现的行，以重复次数开头。该程序将引入if语句，map数据类型以及bufio包。
	lines := make(map[string]int)

	scanner := bufio.NewScanner(os.Stdin)

	strChan := make(chan string)

	go func() {
		for scanner.Scan() {
			// lines[scanner.Text()]++
			strChan <- scanner.Text()
		}
		close(strChan)
	}()

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
	defer cancel()

	for {
		select {
		case str, ok := <-strChan:
			if !ok {
				// 输入结束
				printResults(lines)
				return
			}
			lines[str]++
		case <-ctx.Done():
			printResults(lines)
			fmt.Println("超时")
			return
		}
	}

}

func printResults(lines map[string]int) {
	for line, count := range lines {
		if count > 1 {
			fmt.Printf("%d\t%s\n", count, line)
		}

	}
}
