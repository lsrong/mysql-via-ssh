package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
)

type Conf struct {
	SSH   sshConf   `mapstructure:"ssh"`
	Mysql mysqlConf `mapstructure:"mysql"`
}

type sshConf struct {
	Host string `mapstructure:"host"`
	User string `mapstructure:"user"`
	Pwd  string `mapstructure:"pwd"`
	Port int    `mapstructure:"port"`
}

type mysqlConf struct {
	Host     string `mapstructure:"host"`
	User     string `mapstructure:"user"`
	Pwd      string `mapstructure:"pwd"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
}

type SSHDialer struct {
	client *ssh.Client
}

func (s *SSHDialer) Dial(_ context.Context, addr string) (net.Conn, error) {
	return s.client.Dial("tcp", addr)
}

func (s *SSHDialer) Close() error {
	return s.client.Close()
}

var (
	configFile string
	Config     *Conf
)

func main() {
	rootCmd := cobra.Command{
		Use:           "mysql-via-ssh",
		Short:         "Connect to mysql database through ssh agent",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			flags.StringVarP(&configFile, "config", "c", getDefConfigFile(), "Configuration File")
			return run()
		},
	}

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("rootCmd failed: %v", err)
	}
}

func run() error {
	if err := loadConf(configFile); err != nil {
		return err
	}

	sshDialer, err := openSSH()

	if err != nil {
		return fmt.Errorf("openSSH failed: %v", err)
	}
	defer sshDialer.Close()

	mysql.RegisterDialContext("mysql+tcp", sshDialer.Dial)

	db, err := openMysql()
	if err != nil {
		return fmt.Errorf("openMysql failed: %v", err)
	}
	defer db.Close()

	// todo SQL processing logic
	rows, err := db.Query("SELECT id, name FROM `table` ORDER BY id")
	if err != nil {
		return fmt.Errorf("db.Query failed: %v", err)
	}

	for rows.Next() {
		var id int64
		var name string
		_ = rows.Scan(&id, &name)
		fmt.Printf("ID: %d  Name: %s\n", id, name)
	}
	defer rows.Close()

	return nil
}

func getDefConfigFile() string {
	pwd, _ := os.Getwd()

	return fmt.Sprintf("%s/config.toml", pwd)
}

func loadConf(fpath string) error {
	viper.SetConfigFile(fpath)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return viper.Unmarshal(&Config)
}

func openSSH() (*SSHDialer, error) {
	// ssh config
	conf := Config.SSH
	sshConfig := &ssh.ClientConfig{
		User:            conf.User,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}
	if conf.Pwd != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.PasswordCallback(func() (string, error) {
			return conf.Pwd, nil
		}))
	}

	sshCon, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", conf.Host, conf.Port), sshConfig)
	if err != nil {
		return nil, err
	}

	return &SSHDialer{client: sshCon}, nil
}

func openMysql() (*sql.DB, error) {
	conf := Config.Mysql
	dsn := fmt.Sprintf("%s:%s@mysql+tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&allowNativePasswords=true",
		conf.User, conf.Pwd, conf.Host, conf.Port, conf.Database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
