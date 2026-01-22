package excel

import (
	"fmt"
	"github.com/misakacoder/kagome/cond"
	"github.com/xuri/excelize/v2"
	"io"
	"os"
	"reflect"
	"strconv"
)

type Reader[T any] interface {
	Read() (T, error)
}

type column struct {
	index  int
	header string
	width  float64
	align  string
}

func WriteFile[T any](data []T, filename string) error {
	return write(filename, func(f *os.File) error {
		return Write(data, f)
	})
}

func Write[T any](data []T, writer io.Writer) error {
	cols := columns[T]()
	excel := excelize.NewFile()
	defer excel.Close()
	sheetName := excel.GetSheetName(0)
	err := setHeader(excel, sheetName, cols, func(headers []any) error {
		return excel.SetSheetRow(sheetName, "A1", &headers)
	})
	if err != nil {
		return err
	}
	rowNum := 2
	row := make([]any, len(cols))
	for _, item := range data {
		cell, _ := excelize.CoordinatesToCellName(1, rowNum)
		valid, errs := setRow(item, cols, row, func(row []any) error {
			return excel.SetSheetRow(sheetName, cell, &row)
		})
		if errs != nil {
			return errs
		}
		if valid {
			rowNum++
		}
	}
	if err = excel.Write(writer); err != nil {
		return err
	}
	return nil
}

func StreamWriteFile[T any](reader Reader[T], filename string) error {
	return write(filename, func(f *os.File) error {
		return StreamWrite(reader, f)
	})
}

func StreamWrite[T any](reader Reader[T], writer io.Writer) error {
	cols := columns[T]()
	excel := excelize.NewFile()
	defer excel.Close()
	sheetName := excel.GetSheetName(0)
	streamWriter, err := excel.NewStreamWriter(sheetName)
	if err != nil {
		return err
	}
	err = setHeader(excel, sheetName, cols, func(headers []any) error {
		return streamWriter.SetRow("A1", headers)
	})
	if err != nil {
		return err
	}
	rowNum := 2
	row := make([]any, len(cols))
	for {
		data, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		cell, _ := excelize.CoordinatesToCellName(1, rowNum)
		valid, errs := setRow(data, cols, row, func(row []any) error {
			return streamWriter.SetRow(cell, row)
		})
		if errs != nil {
			return errs
		}
		if valid {
			rowNum++
		}
	}
	if err = streamWriter.Flush(); err != nil {
		return err
	}
	if err = excel.Write(writer); err != nil {
		return err
	}
	return nil
}

func write(filename string, writer func(f *os.File) error) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	return writer(f)
}

func columns[T any]() []column {
	var v T
	tp := reflect.TypeOf(v)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	if tp.Kind() != reflect.Struct {
		panic(fmt.Sprintf("type T must be a struct or pointer to struct, got %s", tp.Kind()))
	}
	var cols []column
	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		tag := field.Tag.Get("excel")
		if tag != "" {
			width := 15.0
			if widthString := field.Tag.Get("width"); widthString != "" {
				if w, err := strconv.ParseFloat(widthString, 64); err == nil {
					width = w
				}
			}
			align := field.Tag.Get("align")
			align = cond.Ternary(align != "", align, "left")
			cols = append(cols, column{index: i, header: tag, width: width, align: align})
		}
	}
	if len(cols) == 0 {
		panic("no exportable column")
	}
	return cols
}

func setHeader(excel *excelize.File, sheetName string, cols []column, setHeader func(headers []any) error) error {
	headers := make([]any, len(cols))
	for i, col := range cols {
		headers[i] = col.header
		colName, _ := excelize.ColumnNumberToName(i + 1)
		if err := excel.SetColWidth(sheetName, colName, colName, col.width); err != nil {
			return err
		}
		style, err := excel.NewStyle(&excelize.Style{
			Alignment: &excelize.Alignment{
				Horizontal: col.align,
				Vertical:   "center",
			},
		})
		if err != nil {
			return err
		}
		if err = excel.SetColStyle(sheetName, colName, style); err != nil {
			return err
		}
	}
	return setHeader(headers)
}

func setRow[T any](v T, cols []column, row []any, setRow func(row []any) error) (bool, error) {
	value := reflect.ValueOf(v)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return false, nil
		}
		value = value.Elem()
	}
	for i, col := range cols {
		field := value.Field(col.index)
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				row[i] = ""
			} else {
				row[i] = field.Elem().Interface()
			}
		} else {
			row[i] = field.Interface()
		}
	}
	return true, setRow(row)
}
