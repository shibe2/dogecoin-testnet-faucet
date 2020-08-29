// SPDX-License-Identifier: AGPL-3.0-or-later

package sqldb

import "github.com/mattn/go-sqlite3"

var createSQLite = []string{`CREATE TABLE "claims" (
  "id" INTEGER NOT NULL PRIMARY KEY,
  "time" DATETIME NOT NULL,
  "client" BLOB(16) NOT NULL,
  "recipient" VARCHAR(35) COLLATE BINARY NOT NULL,
  "amount" REAL NOT NULL,
  "txid" BLOB(32) NOT NULL
)`, `CREATE INDEX "claim_time" ON "claims" ("time")`}

func init() {
	sqlite3.SQLiteTimestampFormats = []string{"2006-01-02 15:04:05"}
	CreateSQL["sqlite3"] = createSQLite
}
