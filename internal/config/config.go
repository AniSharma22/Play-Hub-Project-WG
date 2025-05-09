package config

var (
	Host     = "localhost"
	Port     = 5432
	User     = "postgres"
	Password = "password"
	Dbname   = "play-hub-web"
)

const APP_PASSWORD = "jumsoclsrajifgua"
const REGION = "ap-south-1"
const BUCKET = "play-hub"
const PORT = ":8081"

var Avatars = []string{
	"https://play-hub.s3.ap-south-1.amazonaws.com/avatars/avatar1.png",
	"https://play-hub.s3.ap-south-1.amazonaws.com/avatars/avatar2.png",
	"https://play-hub.s3.ap-south-1.amazonaws.com/avatars/avatar3.png",
	"https://play-hub.s3.ap-south-1.amazonaws.com/avatars/avatar4.png",
	"https://play-hub.s3.ap-south-1.amazonaws.com/avatars/avatar5.png",
	"https://play-hub.s3.ap-south-1.amazonaws.com/avatars/avatar6.png",
	"https://play-hub.s3.ap-south-1.amazonaws.com/avatars/avatar7.png",
	"https://play-hub.s3.ap-south-1.amazonaws.com/avatars/avatar8.png",
}

var MY_SIGNING_KEY = []byte("SECRET")
