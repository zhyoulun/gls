package debug

import (
	log "github.com/sirupsen/logrus"
	"os"
)

var Csv *CsvDebug

const folder = "./output/debug/"

type CsvDebug struct {
	csvFiles map[string]*myFile
	ch       chan *Message
}

type myFile struct {
	f              *os.File
	firstLineWrite bool
}

func newCsvDebug() (*CsvDebug, error) {
	return &CsvDebug{
		csvFiles: make(map[string]*myFile),
		ch:       make(chan *Message, 10000),
	}, nil
}

type Message struct {
	FileName   string
	HeaderLine string
	BodyLine   string
	Close      bool
}

func Init() error {
	var err error
	if err = os.MkdirAll(folder, os.ModePerm); err != nil {
		return err
	}
	if Csv, err = newCsvDebug(); err != nil {
		return err
	}
	go Csv.run()
	return nil
}

func (d *CsvDebug) run() {
	for {
		select {
		case m := <-d.ch:
			if m.Close {
				d.doClose(m)
			} else {
				d.doMessage(m)
			}
		}
	}
}

func (d *CsvDebug) doClose(m *Message) {
	var mf *myFile
	var ok bool
	if mf, ok = d.csvFiles[m.FileName]; ok {
		if err := mf.f.Close(); err != nil {
			filePath := folder + m.FileName
			log.Warnf("close file [%s] err: %+v", filePath, err)
		}
	}
}

func (d *CsvDebug) doMessage(m *Message) {
	var f *os.File
	var mf *myFile
	var ok bool
	var err error

	//create file if not exist
	filePath := folder + m.FileName
	if mf, ok = d.csvFiles[m.FileName]; !ok {
		f, err = os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
		if err != nil {
			log.Warnf("os OpenFile [%s] err: %+v", filePath, err)
			return
		}
		mf = &myFile{
			f: f,
		}
		d.csvFiles[m.FileName] = mf
	}

	//写csv header
	if !mf.firstLineWrite {
		_, err = mf.f.WriteString(m.HeaderLine)
		if err != nil {
			log.Warnf("write string [%s] to file [%s] err: %+v", m.HeaderLine, filePath, err)
		}
		mf.firstLineWrite = true
	}

	//写csv body
	_, err = mf.f.WriteString(m.BodyLine)
	if err != nil {
		log.Warnf("write string [%s] to file [%s] err: %+v", m.BodyLine, filePath, err)
	}
}

func (d *CsvDebug) Write(m *Message) {
	select {
	case d.ch <- m:
	default:
	}
}
