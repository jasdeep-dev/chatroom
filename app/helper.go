package app

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func Titleize(name string) string {
	caser := cases.Title(language.English)
	return caser.String(name)
}

func GetMessagesByGroupID(groupID string) []Message {
	var ctx context.Context
	var err error
	var messages []Message

	query := `
		SELECT
			id,
			timestamp,
			text,
			user_id,
			group_id,
			first_name AS name,
			email
		FROM
			messages
		WHERE group_id=$1`

	// GetGroupsViaAPI()
	rows, err := DBConn.Query(ctx, query, groupID)
	if err != nil {
		log.Println("Error GetUsers from Keycloak", err)
		return messages
	}

	messages, err = pgx.CollectRows(rows, pgx.RowToStructByName[Message])
	if err != nil {
		log.Println("Error GetUsers from Keycloak", err)
		return messages
	}
	defer rows.Close()

	return messages
}
