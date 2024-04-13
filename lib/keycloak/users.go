package keycloak

import (
	"chatroom/app"
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

func GetUsers(ctx context.Context) ([]app.KeyCloakUser, error) {
	var err error
	var users []app.KeyCloakUser

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

	users, err = pgx.CollectRows(rows, pgx.RowToStructByName[app.KeyCloakUser])
	defer rows.Close()
	fmt.Println("Users in Keycloak :=", users)
	return users, err
}

func FindUserByID(ctx context.Context, id string) (app.KeyCloakUser, error) {
	var user app.KeyCloakUser

	query := `
		SELECT
			usr.*, json_agg(json_build_object('name', attr.name, 'value', attr.value)) AS attributes
		FROM
			public.user_entity usr
		JOIN
			public.user_attribute attr ON usr.id = attr.user_id
		WHERE
			usr.id = $1
		GROUP BY
			usr.id;
	`

	err := app.KeycloackDBConn.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.EmailConstraint,
		&user.EmailVerified,
		&user.Enabled,
		&user.FederationLink,
		&user.FirstName,
		&user.LastName,
		&user.RealmID,
		&user.Username,
		&user.CreatedTimestamp,
		&user.ServiceAccountClientLink,
		&user.NotBefore,
		&user.Attributes,
	)
	if err != nil {
		return user, fmt.Errorf("error scanning row: %w", err)
	}
	return user, nil
}
