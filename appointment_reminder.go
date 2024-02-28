package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
    // Abrir la conexión a la base de datos
   
    // Bucle infinito para ejecutar el script continuamente
    for {
        // Calcular el tiempo hasta las 16:45
        duracion := duracionHasta(16, 48)

        // Esperar hasta las 16:45
        time.Sleep(duracion)

		db, err := sql.Open("mysql", "root:h1EFA3CDf246eh642fabhBDBgChEEfGH@tcp(viaduct.proxy.rlwy.net:41346)/railway")
		if err != nil {
			panic(err)
		}
		defer db.Close()
	

        // Realizar el trabajo necesario a las 16:45
        ejecutarTarea(db)

		
    }
}

func duracionHasta(hour, minute int) time.Duration {
    now := time.Now()
    targetTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
    if now.After(targetTime) {
        targetTime = targetTime.Add(24 * time.Hour) // si ya pasó hoy, se programa para mañana
    }
    return targetTime.Sub(now)
}

func ejecutarTarea(db *sql.DB) {
    // Obtener la fecha y hora actual
    now := time.Now()

    initDayBefore := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
    endDayBefore := time.Date(now.Year(), now.Month(), now.Day()+1, 23, 59, 59, 0, now.Location())

    fmt.Println(initDayBefore.Format("2006-01-02 15:04:05"))
    fmt.Println(endDayBefore.Format("2006-01-02 15:04:05"))

    // Realizar la consulta a la base de datos para obtener las citas dentro de las próximas 24 horas
    rows, err := db.Query("SELECT appointments.idUser, users.phone, users.email FROM appointments INNER JOIN users ON appointments.idUser = users.idUser WHERE appointments.appointmentDate BETWEEN ? AND ? AND users.email IS NOT NULL AND appointments.canceled != 1", initDayBefore.Format("2006-01-02 15:04:05"), endDayBefore.Format("2006-01-02 15:04:05"))
    if err != nil {
        panic(err)
    }
    defer rows.Close()

    // Iterar sobre los resultados y mostrarlos
    for rows.Next() {
        var idUser int
        var phone string
        var email string
        // Verificar si el campo de correo electrónico no es nulo antes de imprimir
        if err := rows.Scan(&idUser, &phone, &email); err != nil {
            panic(err)
        }

        // Aquí puedes implementar el código para enviar el recordatorio, como enviar un mensaje de texto al número de teléfono o enviar un correo electrónico a la dirección de correo electrónico.
        // Enviar un mensaje de texto al número de teléfono
        // Enviar un correo electrónico a la dirección de correo electrónico
        // Imprimir el resultado
        fmt.Println(idUser, phone, email)
    }
    // Verificar si hubo un error durante la iteración
    if err := rows.Err(); err != nil {
        panic(err)
    }
}
