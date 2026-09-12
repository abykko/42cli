package tui

import (
	"log"
	"time"

	tea "charm.land/bubbletea/v2"

	"42tui/conf"
	deployment "42tui/deployment"
	"42tui/tui/service"
	"42tui/tui/views"
)

type AutologinMsg struct {
	User string
	Pass string
}

type LoginMsg struct{}

type ProfileMsg struct {
	User string
}

type ClockUpdate struct{}

// Async responses to avoid UI freezing
type StartupResultMsg struct {
	Ok  bool
	Err error
}

type LoginResultMsg struct {
	Ok   bool
	User string
	Pass string
	Err  error
}

type ProfileResultMsg struct {
	Profile service.ProfileData
	Err     error
}

// Async commands

func startupCmd() tea.Cmd {
	return func() tea.Msg {
		ok, err := service.Startup()
		return StartupResultMsg{Ok: ok, Err: err}
	}
}

func loginCmd(user, pass string) tea.Cmd {
	return func() tea.Msg {
		ok, err := service.Login(user, pass)
		return LoginResultMsg{Ok: ok, User: user, Pass: pass, Err: err}
	}
}

func fetchProfileCmd(user string) tea.Cmd {
	return func() tea.Msg {
		prof, err := service.Profile(user)
		return ProfileResultMsg{Profile: prof, Err: err}
	}
}

// Helper function
func getDateTime() (string, string) {
	now := time.Now()
	day := now.Format("02 Jan")
	timeStr := now.Format("15:04:05")
	return day, timeStr
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	// Init clock
	if m.Clock.Time == "" {
		m.Clock.Day, m.Clock.Time = getDateTime()
		return m, func() tea.Msg { return ClockUpdate{} }
	}

	var cmds []tea.Cmd

	// Update viewport
	var vpCmd tea.Cmd
	m.ProjectsViewport, vpCmd = m.ProjectsViewport.Update(msg)
	cmds = append(cmds, vpCmd)

	switch msg := msg.(type) {

	case tea.KeyMsg:
		log.Println("KEY:", msg.String())

		if msg.String() == "ctrl+c" {
			log.Println("[System] exiting...")
			return m, tea.Quit
		}

		// TMP DEBUG
		if msg.String() == "a" {
			return m, func() tea.Msg {
				return ProfileMsg{User: "msouiyeh"}
			}
		}
		if msg.String() == "b" {
			return m, func() tea.Msg {
				return ProfileMsg{User: "iamrani-"}
			}
		}

	case ClockUpdate:
		m.Clock.Day, m.Clock.Time = getDateTime()

		// Mantener el reloj tickeando de forma asíncrona
		cmds = append(cmds, func() tea.Msg {
			time.Sleep(1 * time.Second)
			return ClockUpdate{}
		})

	case AutologinMsg:
		m.Page = "autologin"
		// Disparamos la petición de login en segundo plano
		return m, loginCmd(msg.User, msg.Pass)

	case LoginResultMsg:
		// Procesamos el resultado del login cuando el servidor responde
		if msg.Err != nil {
			log.Println("[Autologin] error:", msg.Err)
			return m, tea.Quit
		}

		if !msg.Ok {
			return m, func() tea.Msg { return LoginMsg{} }
		}

		return m, func() tea.Msg {
			return ProfileMsg{User: msg.User}
		}

	case LoginMsg:
		m.Page = "login"

	case ProfileMsg:
		m.Page = "profile" // Estado visual de carga de perfil
		// Disparamos la petición del perfil en segundo plano sin congelar la app
		return m, fetchProfileCmd(msg.User)

	case ProfileResultMsg:
		if msg.Err != nil {
			if msg.Err.Error() == "freeze" {
				log.Println("[Profile] User is currently frozen.")
				m.Profile.ID = -1
				break
			}

			log.Println("[Profile] error:", msg.Err)
			return m, tea.Quit
		}

		m.Profile = msg.Profile
		m.ProjectsViewport.SetContent(views.Projects(m.Profile.Projects))

	case StartupResultMsg:
		// Procesamos el resultado del Startup que se lanzó al inicio
		if msg.Err != nil {
			deployment.Stop()
			return m, tea.Quit
		}

		if msg.Ok {
			autoLogin, err := conf.GetBool("autologin")
			if err != nil {
				log.Println("[Startup] autologin config error:", err)
				return m, tea.Quit
			}

			if !autoLogin {
				return m, func() tea.Msg { return LoginMsg{} }
			}

			user, pass, _ := service.Autologin()

			return m, func() tea.Msg {
				return AutologinMsg{
					User: user,
					Pass: pass,
				}
			}
		}

	case tea.WindowSizeMsg:
		// m.ProjectsViewport.Width = msg.Width
		// m.ProjectsViewport.Height = msg.Height
	}

	// Lógica de Startup delegada a un comando asíncrono
	if m.Page == "startup" {
		m.Page = "loading_startup" // Cambiamos el estado temporalmente para no repetir este IF
		return m, startupCmd()
	}

	return m, tea.Batch(cmds...)
}