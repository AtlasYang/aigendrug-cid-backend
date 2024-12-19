package app

import (
	"archive/tar"
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func WeightKey(jobID int) string {
	return fmt.Sprintf("weight-%d.pth", jobID)
}

func AGKey(jobID int) string {
	return fmt.Sprintf("ag-job-%d.tar", jobID)
}

func AGPath(jobID int) string {
	return fmt.Sprintf("/app/weights/ag-job-%d", jobID)
}

func ExtractTar(tarball []byte, outputDir string) error {
	tarReader := tar.NewReader(bytes.NewReader(tarball))

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading tar entry: %w", err)
		}

		targetPath := filepath.Join(outputDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			err := os.MkdirAll(targetPath, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("error creating directory: %w", err)
			}
		case tar.TypeReg:
			err := os.MkdirAll(filepath.Dir(targetPath), 0755)
			if err != nil {
				return fmt.Errorf("error creating file directory: %w", err)
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				return fmt.Errorf("error creating file: %w", err)
			}

			_, err = io.Copy(outFile, tarReader)
			if err != nil {
				outFile.Close()
				return fmt.Errorf("error writing to file: %w", err)
			}

			outFile.Close()
		default:
			fmt.Printf("Ignoring unsupported type: %c in file %s\n", header.Typeflag, header.Name)
		}
	}
	return nil
}

func ArchiveTar(inputDir string) ([]byte, error) {
	buf := new(bytes.Buffer)
	tarWriter := tar.NewWriter(buf)

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking path: %w", err)
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return fmt.Errorf("error creating file info header: %w", err)
		}

		header.Name = path

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("error writing header: %w", err)
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening file: %w", err)
			}

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				file.Close()
				return fmt.Errorf("error copying file: %w", err)
			}

			file.Close()
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	err = tarWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing tar writer: %w", err)
	}

	return buf.Bytes(), nil
}

func LigandDataToCsv(data []LigandData, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	defer file.Close()

	_, err = file.WriteString("smiles,standard_value\n")
	if err != nil {
		return fmt.Errorf("error writing header: %w", err)
	}

	for _, ligand := range data {
		_, err = file.WriteString(fmt.Sprintf("%s,%f\n", ligand.SMILES, ligand.StdValue))
		if err != nil {
			return fmt.Errorf("error writing ligand data: %w", err)
		}
	}

	return nil
}

func CsvToLigandData(path string) ([]LigandData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	defer file.Close()

	var data []LigandData

	// Skip header
	csvReader := csv.NewReader(bufio.NewReader(file))
	_, err = csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading csv header: %w", err)
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading csv record: %w", err)
		}

		sv, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, fmt.Errorf("error parsing standard value: %w", err)
		}

		ligand := LigandData{
			SMILES:   record[0],
			StdValue: sv,
		}

		data = append(data, ligand)
	}

	return data, nil
}
