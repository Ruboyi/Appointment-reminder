package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	_ "github.com/go-sql-driver/mysql"
)

var (
	apiKey            string
	sqlAlchemyURI     string
    targetHour        string
    targetMinute      string
	timezone          string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey = os.Getenv("API_KEY_SENDGRID")
	sqlAlchemyURI = os.Getenv("SQL_ALCHEMY_DATABASE_URI")
    targetHour = os.Getenv("TARGET_HOUR")
    targetMinute = os.Getenv("TARGET_MINUTE")
	timezone = os.Getenv("TIMEZONE")
}

func sendEmail(toEmail, name, appointmentDate string) error {
	from := mail.NewEmail("Atkinson Barber Shop", "Atkinsonbarbershop@gmail.com")
	subject := "Recordatorio de Cita: ¡Te Esperamos Pronto!"
	to := mail.NewEmail("Estimado/a "+name, toEmail)

	appointmentDateTime, err := time.Parse("2006-01-02 15:04:05", appointmentDate)
	if err != nil {
		return err
	}
	appointmentTime := appointmentDateTime.Format("15:04")

	htmlContent := fmt.Sprintf(`<div style="text-align: center;">
    <img src="https://atkinsonbarbershop.com/wp-content/uploads/2017/06/logoatkinsonheader.png" alt="Logo Atkinson Barber Shop" style="width: 200px; height: auto; margin: 20px auto;">
    <h1>Atkinson Barber Shop</h1>
    <p>Hola %s,</p>
    <p>Te recordamos que tienes una cita programada con nosotros para mañana:</p>
    <p>Hora de la cita: %s</p>
    <p>Por favor, asegúrate de llegar a tiempo. Estamos ansiosos por verte.</p>
    <p>Si necesitas reprogramar o cancelar tu cita puedes hacerlo a través de nuestra App, por favor contáctanos con anticipación.</p>
    <p>¡Gracias y nos vemos pronto en Atkinson Barber Shop!</p>
  </div>`, name, appointmentTime)

	message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
	client := sendgrid.NewSendClient(apiKey)
	response, err := client.Send(message)
	if err != nil {
		return err
	}
	fmt.Println(response.StatusCode)
	fmt.Println(response.Body)
	fmt.Println(response.Headers)
	return nil
}

func main() {
	fmt.Println("Iniciando el servicio de recordatorio de citas...")
    fmt.Println("Hora objetivo: ", targetHour + ":" + targetMinute)
	for {
        fmt.Println("Esperando a la hora objetivo...")
		now := time.Now().In(time.FixedZone(timezone, getTimezoneOffset(timezone)))
		
        targetTimeParts := []int{0, 0}
        fmt.Sscanf(targetHour, "%d", &targetTimeParts[0])
        fmt.Sscanf(targetMinute, "%d", &targetTimeParts[1])

        targetTime := time.Date(now.Year(), now.Month(), now.Day(), targetTimeParts[0], targetTimeParts[1], 0, 0, now.Location())
        fmt.Println("Target time: ", targetTime)
		if now.After(targetTime) {
			targetTime = targetTime.Add(24 * time.Hour)
		}
		duration := targetTime.Sub(now)

        fmt.Println("Esperando ", duration)
		time.Sleep(duration)

        fmt.Println("Hora de enviar correos")
        fmt.Println("Conectando a la base de datos...")
		db, err := sql.Open("mysql", sqlAlchemyURI)
		if err != nil {
			panic(err)
		}
        fmt.Println("Conexión exitosa")
		now = time.Now().In(time.FixedZone(timezone, getTimezoneOffset(timezone)))
		initDayBefore := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		endDayBefore := time.Date(now.Year(), now.Month(), now.Day()+1, 23, 59, 59, 0, now.Location())

		rows, err := db.Query("SELECT appointments.idUser, appointments.appointmentDate, users.nameUser, users.email FROM appointments INNER JOIN users ON appointments.idUser = users.idUser WHERE appointments.appointmentDate BETWEEN ? AND ? AND users.email IS NOT NULL AND appointments.canceled != 1", initDayBefore.Format("2006-01-02 15:04:05"), endDayBefore.Format("2006-01-02 15:04:05"))
		if err != nil {
			panic(err)
		}

		for rows.Next() {
			var idUser int
			var appointmentDate string
			var name string
			var email string

			if err := rows.Scan(&idUser, &appointmentDate, &name, &email); err != nil {
				panic(err)
			}
			fmt.Println("Enviando correo a: ", email)
			sendEmail(email, name, appointmentDate)

			time.Sleep(10 * time.Second)
		}

        rows.Close()

		db.Close()
		if err := rows.Err(); err != nil {
			panic(err)
		}

        fmt.Println("Correos enviados")
	}

}

func getTimezoneOffset(zone string) int {
	_, offset := time.Now().In(time.FixedZone(zone, 0)).Zone()
	return offset
}

