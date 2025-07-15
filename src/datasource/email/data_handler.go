// handler.go
package email

import (
	"bytes"
	"encoding/csv"
	"sync"

	"github.com/go-gota/gota/dataframe"
	"github.com/tealeg/xlsx"
)

// DataFrameWrapper 封装DataFrame并提供线程安全访问
type DataFrameWrapper struct {
	df dataframe.DataFrame // 存储DataFrame数据
	mu sync.RWMutex        // 读写锁保证线程安全
}

// LoadCSV 从CSV数据加载到DataFrame
func (d *DataFrameWrapper) LoadCSV(data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	r := csv.NewReader(bytes.NewReader(data))
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	d.df = dataframe.LoadRecords(records)
	return nil
}

// LoadXLSX 从XLSX数据加载到DataFrameZ
func (d *DataFrameWrapper) LoadXLSX(data []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	xlFile, err := xlsx.OpenBinary(data)
	if err != nil {
		return err
	}

	var records [][]string
	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			var record []string
			for _, cell := range row.Cells {
				record = append(record, cell.String())
			}
			records = append(records, record)
		}
	}

	d.df = dataframe.LoadRecords(records)
	return nil
}

// GetDF 获取当前DataFrame(线程安全)
func (d *DataFrameWrapper) GetDF() dataframe.DataFrame {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.df
}
