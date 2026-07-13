package config

type AppConfig struct {
	Mysql
	Redis
	Elastic
	Sms
	AliSms
	SmRz
	Qny
	Alipay
	RabbitMq
	KafkaMq
	AI
	JWT
}
type Mysql struct {
	User     string
	Password string
	Host     string
	Port     int
	Database string
}
type Redis struct {
	Host     string
	Port     int
	Password string
	Database int
}
type Elastic struct {
	Host string
}
type Sms struct {
	APIID  string
	APIKEY string
}
type AliSms struct {
	AccessKeyId     string
	AccessKeySecret string
	SignName        string
	TemplateCode    string
}
type SmRz struct {
	SecretID  string
	SecretKey string
	UseID     string
}
type Qny struct {
	AccessKey string
	SecretKey string
	Bucket    string
	Domain    string
}
type Alipay struct {
	AppId      string
	PrivateKey string
	NotifyUrl  string
	AliPubKey  string
}
type RabbitMq struct {
	Host string
}
type KafkaMq struct {
	Host string
}
type AI struct {
	APIKey  string
	Model   string
	BaseURL string
}
type JWT struct {
	AppKey string
}
