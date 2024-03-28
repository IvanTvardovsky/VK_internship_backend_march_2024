package structures

type Config struct {
	Listen  Listener      `yaml:"listen"`
	Storage StorageConfig `yaml:"storage"`
	Key     JWTSecretKey  `yaml:"authorization"`
	Token   TokenStruct   `yaml:"token"`
}

type Listener struct {
	BindIp string `yaml:"bind_ip"`
	Port   string `yaml:"port"`
}

type StorageConfig struct {
	Host     string `yaml:"host"`
	Port     rune   `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type JWTSecretKey struct {
	SecretKey string `yaml:"key"`
}

type TokenStruct struct {
	Expires string `yaml:"expires"`
}

type TokenInfo struct {
	Login   string `json:"login"`
	ID      int    `json:"user_id"`
	Expires int    `json:"expires"`
}

type User struct {
	ID           int    `json:"id"`
	Login        string `json:"login"`
	Password     string `json:"password"`
	PasswordHash string `json:"passwordHash"`
}

type Ad struct {
	ID           int    `json:"id"`
	UserID       int    `json:"user_id"`
	AuthorLogin  string `json:"author_login"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ImageAddress string `json:"image_address"`
	Price        int    `json:"price"`
	CreatedAt    string `json:"created_at"`
	IsOwner      bool   `json:"is_owner"`
}
