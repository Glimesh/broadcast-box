package migrations

import (
	"encoding/json"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/daos"
	m "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/models"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		jsonData := `{
			"id": "bl7jh9ptmhrj95q",
			"created": "2024-10-05 21:03:37.812Z",
			"updated": "2024-10-05 21:03:37.812Z",
			"name": "streamkeys",
			"type": "base",
			"system": false,
			"schema": [
				{
					"system": false,
					"id": "90ugkcyb",
					"name": "streamkey",
					"type": "text",
					"required": false,
					"presentable": false,
					"unique": false,
					"options": {
						"min": null,
						"max": null,
						"pattern": ""
					}
				}
			],
			"indexes": [],
			"listRule": null,
			"viewRule": null,
			"createRule": null,
			"updateRule": null,
			"deleteRule": null,
			"options": {}
		}`

		collection := &models.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return daos.New(db).SaveCollection(collection)
	}, func(db dbx.Builder) error {
		dao := daos.New(db);

		collection, err := dao.FindCollectionByNameOrId("bl7jh9ptmhrj95q")
		if err != nil {
			return err
		}

		return dao.DeleteCollection(collection)
	})
}
