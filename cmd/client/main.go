package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	client "github.com/jmsilvadev/zc/cmd/client/internal"
)

func main() {
	if err := run(flag.CommandLine, os.Args[1:]); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// created to facilitate the tests
func run(flagSet *flag.FlagSet, args []string) error {
	dir := flagSet.String("dir", "", "Directory containing files for upload")
	filesList := flagSet.String("files", "", "Comma-separated list of files for upload")
	serverHost := flagSet.String("host", "http://localhost:5000", "Server host")
	operation := flagSet.String("operation", "upload", "Operation to perform: upload, update or download. Attention: perform an upload will always remove the existent data")
	index := flagSet.Int("index", -1, "Index of the file to download")
	del := flagSet.Bool("delete", true, "If the client can delete the local files after the upload")
	configDir := flagSet.String("config-dir", getDefaultConfigDir(), "Directory to store rootHash and downloaded files")

	flagSet.Parse(args)

	if *operation != "upload" && *operation != "update" && *operation != "download" {
		return fmt.Errorf("invalid operation. Please specify 'upload' or 'update' or 'download' using the -operation parameter")
	}

	err := isDirAvailable(*configDir)
	if err != nil {
		return fmt.Errorf("error creating config directory: %s", err)
	}

	c := client.NewClient(*serverHost)
	if *operation == "upload" {
		err := upload(c, dir, filesList, *configDir)
		if err == nil && *del {
			return removeLocalFiles(*dir, *filesList)
		}
		return nil
	}

	if *operation == "update" {
		err := update(c, dir, filesList, *configDir)
		if err == nil && *del {
			return removeLocalFiles(*dir, *filesList)
		}
		return nil
	}

	return download(c, index, configDir)
}

func upload(c *client.Client, dir, filesList *string, configDir string) error {
	if *dir == "" && *filesList == "" {
		return fmt.Errorf("please provide the directory containing the files using the -dir parameter or a list of files using the -files parameter")
	}

	var files [][]byte
	if *dir != "" {
		files = getFilesFromDir(*dir)
	}

	if *filesList != "" {
		files = getFiles(*filesList)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found for upload")
	}

	rootHash := c.GetRootHash(files)
	_, err := c.UploadFiles(files)
	if err != nil {
		return fmt.Errorf("error uploading files: %s", err)
	}

	if !isValid(c, files, rootHash) {
		return fmt.Errorf("the upload process was unsuccessful, it's not safe to delete from the local filesystem")
	}

	rootHashPath := filepath.Join(configDir, ".rootHash")
	err = os.WriteFile(rootHashPath, []byte(rootHash), 0644)
	if err != nil {
		return fmt.Errorf("error saving rootHash: %s error: %s", rootHashPath, err)
	}

	fmt.Println("All files were uploaded and validated properly.")
	return nil
}

func update(c *client.Client, dir, filesList *string, configDir string) error {
	if *dir == "" && *filesList == "" {
		return fmt.Errorf("please provide the directory containing the files using the -dir parameter or a list of files using the -files parameter")
	}

	var files [][]byte
	if *dir != "" {
		files = getFilesFromDir(*dir)
	}

	if *filesList != "" {
		files = getFiles(*filesList)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found for upload")
	}

	rootHash, err := c.UpdateFiles(files, configDir)
	if err != nil {
		return fmt.Errorf("error uploading files: %s", err)
	}

	if !isValid(c, files, rootHash) {
		return fmt.Errorf("the upload process was unsuccessful, it's not safe to delete from the local filesystem")
	}

	rootHashPath := filepath.Join(configDir, ".rootHash")
	err = os.WriteFile(rootHashPath, []byte(rootHash), 0644)
	if err != nil {
		return fmt.Errorf("error saving rootHash: %s error: %s", rootHashPath, err)
	}

	fmt.Println("All files were uploaded and validated properly.")
	return nil
}

func download(c *client.Client, index *int, configDir *string) error {
	if *index == -1 || *configDir == "" {
		return fmt.Errorf("please provide the index and configDir parameters for the download operation")
	}

	rootHash, err := c.GetLocalRootHash(*configDir)
	if err != nil {
		return fmt.Errorf("error fetching the rootHash: %s", err)
	}

	file, proof, err := c.DownloadFile(*index, rootHash)
	if err != nil {
		return fmt.Errorf("error downloading file: %s", err)
	}

	if c.VerifyProof(file, proof, rootHash) {
		// TODO: put this filepath as a config env
		filePath := fmt.Sprintf(*configDir+"/downloaded_file_%d", index)
		err = os.WriteFile(filePath, file, 0644)
		if err != nil {
			return fmt.Errorf("error saving file: %s error: %v", filePath, err)
		}

		fmt.Printf("File downloaded, verified and saved as %s\n", filePath)
		return nil
	}

	fmt.Println("The download process was unsuccessful or the file is invalid")
	return nil
}

func isValid(c *client.Client, files [][]byte, rootHash string) bool {
	isValid := true
	for i := range files {
		file, proof, err := c.DownloadFile(i, rootHash)
		if err != nil {
			fmt.Println("Error downloading file:", err)
			return false
		}

		if !c.VerifyProof(file, proof, rootHash) {
			isValid = false
		}
	}
	return isValid
}

func getFiles(filesList string) [][]byte {
	var files [][]byte

	filePaths := strings.Split(filesList, ",")
	for _, filePath := range filePaths {
		filePath = strings.TrimSpace(filePath)
		if filePath == "" {
			continue
		}

		absPath, err := filepath.Abs(filePath)
		if err != nil {
			fmt.Println("Error resolving file path:", filePath, err)
			return nil
		}

		fileContent, err := os.ReadFile(absPath)
		if err != nil {
			fmt.Println("Error reading file:", absPath, err)
			return nil
		}
		files = append(files, fileContent)
	}

	return files
}

func getFilesFromDir(dir string) [][]byte {
	var files [][]byte

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileContent, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			files = append(files, fileContent)
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error reading directory:", err)
		return nil
	}
	return files
}

func getDefaultConfigDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}
	// TODO: put this dirname as a config env
	return filepath.Join(homeDir, ".zc")
}

func isDirAvailable(dir string) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	return nil
}

func removeLocalFiles(dir, filesList string) error {
	if filesList != "" {
		filePaths := strings.Split(filesList, ",")
		for _, filePath := range filePaths {
			filePath = strings.TrimSpace(filePath)
			if filePath == "" {
				continue
			}
			absPath, err := filepath.Abs(filePath)
			if err != nil {
				return fmt.Errorf("error resolving file path: %v", err)
			}
			err = os.Remove(absPath)
			if err != nil {
				return fmt.Errorf("error removing file: %v", err)
			}
		}
	}

	if dir != "" {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				err = os.Remove(path)
				if err != nil {
					return fmt.Errorf("error removing file: %v", err)
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	fmt.Println("All local files have been removed successfully")
	return nil
}
