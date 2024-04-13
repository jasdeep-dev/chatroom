package keycloak

import (
	"chatroom/app"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func GetUsers(ctx context.Context) ([]app.KeyCloackUser, error) {
	var err error
	var users []app.KeyCloackUser

	query := `
	SELECT
		usr.*, json_agg(json_build_object('name', attr.name, 'value', attr.value)) AS attributes
	FROM
		public.user_entity usr
	JOIN
		public.user_attribute attr ON usr.id = attr.user_id
	GROUP BY
		usr.id;
	`

	rows, err := app.KeycloackDBConn.Query(ctx, query)
	if err != nil {
		log.Println("Error GetUsers from Keycloak", err)
		return users, err
	}

	users, err = pgx.CollectRows(rows, pgx.RowToStructByName[app.KeyCloackUser])
	defer rows.Close()
	fmt.Println("Users in Keycloak :=", users)
	return users, err
}
