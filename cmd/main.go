package main

import (

	"fmt"
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/gorilla/mux"

)

func initClient() *mongo.Client {

	var err error
	var client *mongo.Client

	uri := //mongodb uri goes here
	options := options.Client()
	options.ApplyURI(uri)

	if client, err = mongo.Connect(context.Background(), options); err != nil {
		fmt.Println(err.Error())
	}

	return client
}

func initRouter() {

	route := mux.NewRouter()

	route.HandleFunc("/upload", getUpload).Methods("POST")
	route.HandleFunc("/download", getDownload).Methods("GET")
	
	fmt.Println("Server running...")
	log.Fatal(http.ListenAndServe("127.0.0.1:9000", route))
}

func uploadFile(file, filename string) {
	
	data, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	connection := initClient()
	fs, err := gridfs.NewBucket(
		connection.Database("rpi2-storage-db"),
	)

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	upStream, err := fs.OpenUploadStream(
		filename,
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer upStream.Close()

	fileSize, err := upStream.Write(data)

	for bytes := range data {
		fmt.Printf("%d", bytes)
	}

	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	log.Printf("Write to db was successful. File size: %d MBs\n", fileSize)
}

func downloadFile(fileName string) {
	
	connection := initClient()

	dataBase := connection.Database("rpi2-storage-db")
	fsFiles := dataBase.Collection("fs.files")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	var results bson.M
	err := fsFiles.FindOne(ctx, bson.M{}).Decode(&results)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(results)

	fs, _ := gridfs.NewBucket(
		dataBase,
	)

	var buf bytes.Buffer
	downStream, err := fs.DownloadToStreamByName(fileName, &buf)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("File size to download: %v\n", downStream)
	ioutil.WriteFile(fileName, buf.Bytes(), 0600)

}

func getUpload(write http.ResponseWriter, read *http.Request) {
	
	file := os.Args[1]
	filePath := path.Base(file)

	write.Header().Set("Access-Control-Allow-Origin", "*")
    write.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	write.Header().Set("Content-type", "mpeg/mp3")
	uploadFile(file, filePath)
}

func getDownload(write http.ResponseWriter, read *http.Request) {
	
	file := os.Args[1]
	filePath := path.Base(file)

	write.Header().Set("Access-Control-Allow-Origin", "*")
    write.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	write.Header().Set("Content-Type", "mpeg/mp3")
	downloadFile(filePath)
}

func main() {

	initRouter()

}