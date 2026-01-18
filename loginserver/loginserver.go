package loginserver

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"

	"github.com/frostwind/l2go/config"
	"github.com/frostwind/l2go/loginserver/clientpackets"
	"github.com/frostwind/l2go/loginserver/models"
	"github.com/frostwind/l2go/loginserver/serverpackets"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type LoginServer struct {
	clients             []*models.Client
	gameservers         []*models.GameServer
	database            *sql.DB
	config              config.ConfigObject
	internalServersList []byte
	externalServersList []byte
	status              loginServerStatus
	clientsListener     net.Listener
	gameServersListener net.Listener
}

type loginServerStatus struct {
	successfulAccountCreation uint32
	failedAccountCreation     uint32
	successfulLogins          uint32
	failedLogins              uint32
	hackAttempts              uint32
}

func New(cfg config.ConfigObject) *LoginServer {
	return &LoginServer{config: cfg}
}

func (l *LoginServer) Init() {
	var err error

	// Connect to MySQL database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		l.config.LoginServer.Database.User,
		l.config.LoginServer.Database.Password,
		l.config.LoginServer.Database.Host,
		l.config.LoginServer.Database.Port,
		l.config.LoginServer.Database.Name)

	l.database, err = sql.Open("mysql", dsn)
	if err != nil {
		panic("Couldn't connect to the database server: " + err.Error())
	}

	// Test the connection
	err = l.database.Ping()
	if err != nil {
		panic("Couldn't ping the database server: " + err.Error())
	}

	fmt.Println("Successfully connected to the MySQL database server")

	// Listen for client connections
	l.clientsListener, err = net.Listen("tcp", ":2106")
	if err != nil {
		fmt.Println("Couldn't initialize the Login Server (Clients listener)")
	} else {
		fmt.Println("Login Server listening for clients connections on port 2106")
	}

	// Listen for game servers connections
	l.gameServersListener, err = net.Listen("tcp", ":9413")
	if err != nil {
		fmt.Println("Couldn't initialize the Login Server (Gameservers listener)")
	} else {
		fmt.Println("Login Server listening for gameservers connections on port 9413")
	}
}

func (l *LoginServer) Start() {
	defer l.database.Close()
	defer l.clientsListener.Close()
	defer l.gameServersListener.Close()

	done := make(chan bool)

	go func() {
		for {
			var err error
			client := models.NewClient()
			client.Socket, err = l.clientsListener.Accept()
			l.clients = append(l.clients, client)
			if err != nil {
				fmt.Println("Couldn't accept the incoming connection.")
				continue
			} else {
				go l.handleClientPackets(client)
			}
		}
	}()

	go func() {
		for {
			var err error
			gameserver := models.NewGameServer()
			gameserver.Socket, err = l.gameServersListener.Accept()
			l.gameservers = append(l.gameservers, gameserver)
			if err != nil {
				fmt.Println("Couldn't accept the incoming connection.")
				continue
			} else {
				go l.handleGameServerPackets(gameserver)
			}
		}
	}()

	for i := 0; i < 2; i++ {
		<-done
	}

}

func (l *LoginServer) kickClient(client *models.Client) {
	client.Socket.Close()

	for i, item := range l.clients {
		if bytes.Equal(item.SessionID, client.SessionID) {
			copy(l.clients[i:], l.clients[i+1:])
			l.clients[len(l.clients)-1] = nil
			l.clients = l.clients[:len(l.clients)-1]
			break
		}
	}

	fmt.Println("The client has been successfully kicked from the server.")
}

func (l *LoginServer) handleGameServerPackets(gameserver *models.GameServer) {
	defer gameserver.Socket.Close()

	for {
		opcode, _, err := gameserver.Receive()

		if err != nil {
			fmt.Println(err)
			fmt.Println("Closing the connection...")
			break
		}

		switch opcode {
		case 00:
			fmt.Println("A game server sent a request to register")
		default:
			fmt.Println("Can't recognize the packet sent by the gameserver")
		}
	}
}

func (l *LoginServer) handleClientPackets(client *models.Client) {
	fmt.Println("A client is trying to connect...")
	defer l.kickClient(client)

	buffer := serverpackets.NewInitPacket()
	err := client.Send(buffer, false, false)

	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("Init packet sent.")
	}

	for {
		opcode, data, err := client.Receive()

		if err != nil {
			fmt.Println(err)
			fmt.Println("Closing the connection...")
			break
		}

		switch opcode {
		case 00:
			// response buffer
			var buffer []byte

			requestAuthLogin := clientpackets.NewRequestAuthLogin(data)

			fmt.Printf("User %s is trying to login\n", requestAuthLogin.Username)

			// Query for existing account
			var account models.Account
			err := l.database.QueryRow("SELECT id, username, password, access_level FROM accounts WHERE username = ?", requestAuthLogin.Username).Scan(
				&account.Id, &account.Username, &account.Password, &account.AccessLevel)

			if err == sql.ErrNoRows {
				if l.config.LoginServer.AutoCreate == true {
					hashedPassword, err := bcrypt.GenerateFromPassword([]byte(requestAuthLogin.Password), 10)
					if err != nil {
						fmt.Println("An error occured while trying to generate the password")
						l.status.failedAccountCreation += 1

						buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_SYSTEM_ERROR)
					} else {
						// Insert new account
						result, err := l.database.Exec("INSERT INTO accounts (username, password, access_level) VALUES (?, ?, ?)",
							requestAuthLogin.Username, string(hashedPassword), ACCESS_LEVEL_PLAYER)

						if err != nil {
							fmt.Printf("Couldn't create an account for the user %s: %v\n", requestAuthLogin.Username, err)
							l.status.failedAccountCreation += 1

							buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_SYSTEM_ERROR)
						} else {
							accountId, _ := result.LastInsertId()
							client.Account = models.Account{
								Id:          accountId,
								Username:    requestAuthLogin.Username,
								Password:    string(hashedPassword),
								AccessLevel: ACCESS_LEVEL_PLAYER}

							fmt.Printf("Account successfully created for the user %s\n", requestAuthLogin.Username)
							l.status.successfulAccountCreation += 1

							buffer = serverpackets.NewLoginOkPacket(client.SessionID)
						}
					}
				} else {
					fmt.Println("Account not found !")
					l.status.failedLogins += 1

					buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_USER_OR_PASS_WRONG)
				}
			} else if err != nil {
				fmt.Printf("Database error: %v\n", err)
				buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_SYSTEM_ERROR)
			} else {
				// Account exists; Is the password ok?
				client.Account = account
				err = bcrypt.CompareHashAndPassword([]byte(client.Account.Password), []byte(requestAuthLogin.Password))

				if err != nil {
					fmt.Printf("Wrong password for the account %s\n", requestAuthLogin.Username)
					l.status.failedLogins += 1

					buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_USER_OR_PASS_WRONG)
				} else {

					if client.Account.AccessLevel >= ACCESS_LEVEL_PLAYER {
						l.status.successfulLogins += 1

						buffer = serverpackets.NewLoginOkPacket(client.SessionID)
					} else {
						l.status.failedLogins += 1

						buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_ACCESS_FAILED)
					}

				}
			}

			err = client.Send(buffer)

			if err != nil {
				fmt.Println(err)
			}

		case 02:
			requestPlay := clientpackets.NewRequestPlay(data)

			fmt.Printf("The client wants to connect to the server : %d\n", requestPlay.ServerID)

			var buffer []byte
			if len(l.config.GameServers) >= int(requestPlay.ServerID) && (l.config.GameServers[requestPlay.ServerID-1].Options.Testing == false || client.Account.AccessLevel > ACCESS_LEVEL_PLAYER) {
				if !bytes.Equal(client.SessionID[:8], requestPlay.SessionID) {
					l.status.hackAttempts += 1

					buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_ACCESS_FAILED)
				} else {
					buffer = serverpackets.NewPlayOkPacket()
				}
			} else {
				l.status.hackAttempts += 1

				buffer = serverpackets.NewPlayFailPacket(serverpackets.REASON_ACCESS_FAILED)
			}
			err := client.Send(buffer)

			if err != nil {
				fmt.Println(err)
			}

		case 05:
			requestServerList := clientpackets.NewRequestServerList(data)

			var buffer []byte
			if !bytes.Equal(client.SessionID[:8], requestServerList.SessionID) {
				l.status.hackAttempts += 1

				buffer = serverpackets.NewLoginFailPacket(serverpackets.REASON_ACCESS_FAILED)
			} else {
				buffer = serverpackets.NewServerListPacket(l.config.GameServers, client.Socket.RemoteAddr().String())
			}
			err := client.Send(buffer)

			if err != nil {
				fmt.Println(err)
			}

		default:
			fmt.Println("Couldn't detect the packet type.")
		}
	}
}
