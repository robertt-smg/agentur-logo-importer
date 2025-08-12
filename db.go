package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/glog"
)

func (t *this) connectDB() (db *sql.DB, err error) {
	connectionString := t.cfg.Section("db").Key("connectionString").String()
	db, err = sql.Open("mysql", connectionString)

	return db, err
}

/*
REPLACE INTO `tblmok_hist` (`histdate`, `histuser`, `Kdnr`, `MOK`, `import_unique`, `Aktion`)
SELECT

	now() as histdate,

'rtu_logo' as histuser,
fti_importi5umsatz_kdnr.Kdnr,
'A-LOGO',
CONVERT(CONCAT('9',fti_importi5umsatz_kdnr.kdnr, '9212'), UNSIGNED INTEGER) as import_unique,
'1'
from custsupp
JOIN fti_importi5umsatz_kdnr ON fti_importi5umsatz_kdnr.i5nr = custsupp.ccusuppno_ AND fti_importi5umsatz_kdnr.Firma = custsupp.ccompany_
left join tblmok on tblmok.Kdnr = fti_importi5umsatz_kdnr.Kdnr AND tblmok.MOK = 'A-LOGO'

WHERE
clogo_ <> ”
and
tblmok.Kdnr is null;

REPLACE INTO `tblmok` (`Kdnr`, `MOK`, `import_unique`)
SELECT

fti_importi5umsatz_kdnr.Kdnr,
'A-LOGO',
CONVERT(CONCAT('9',fti_importi5umsatz_kdnr.kdnr, '9212'), UNSIGNED INTEGER) as import_unique

from custsupp
JOIN fti_importi5umsatz_kdnr ON fti_importi5umsatz_kdnr.i5nr = custsupp.ccusuppno_ AND fti_importi5umsatz_kdnr.Firma = custsupp.ccompany_
left join tblmok on tblmok.Kdnr = fti_importi5umsatz_kdnr.Kdnr AND tblmok.MOK = 'A-LOGO'

WHERE
clogo_ <> ”
and
tblmok.Kdnr is null;
*/
func (t *this) updateDB(logos map[string]*logo) {

	glog.V(2).Infoln("-> ", logGetCurrentFuncName())

	db, err := t.connectDB()
	if err == nil {
		defer db.Close()
	}
	if isSuccessOrLogError(err, "Cannot Open database") == nil && db != nil {
		for _, logo := range logos {

			if logo.Downloaded {
				glog.Infoln(logGetCurrentFuncName(), logo.Agency, logo.Country, logo.Serverpath)
				// tblmok_hist
				var stmt1 *sql.Stmt
				stmt1, logo.Err = db.Prepare("CALL add_agency_logo(?) ")
				if err == nil {
					defer stmt1.Close()
				}
				if isSuccessOrLogError(logo.Err, "Cannot create A-LOGO Link in Airna") == nil && stmt1 != nil {
					_, logo.Err = stmt1.Exec(logo.Agency)
					logo.Kimed = true
				}
			}
		}
	}
	glog.V(2).Infoln("<- ", logGetCurrentFuncName())

	db.Close()
}
func (t *this) loadSubagents(logo *logo) (ret []string) {
	glog.V(2).Infoln("-> ", logGetCurrentFuncName())

	db, err := t.connectDB()

	if err == nil {
		defer db.Close()
	}
	if isSuccessOrLogError(err, "Cannot Open database") == nil && db != nil {
		glog.Infoln(logGetCurrentFuncName(), logo.Agency, logo.Country, logo.Serverpath)
		// tblmok_hist
		var stmt1 *sql.Stmt
		stmt1, logo.Err = db.Prepare("SELECT " +
			" user " +
			"from fti_importi5umsatz_kdnr " +
			"join fti_subagents ON fti_subagents.kdnr = fti_importi5umsatz_kdnr.kdnr " +
			"WHERE 1 " +
			"AND fti_importi5umsatz_kdnr.i5nr = ?")
		if err == nil {
			defer stmt1.Close()
		}
		if isSuccessOrLogError(logo.Err, "Cannot SELECT from fti_subagents") == nil && stmt1 != nil {
			rows, err := stmt1.Query(logo.Agency)
			if err == nil {
				defer rows.Close()
				ret = make([]string, 0)
				for rows.Next() {
					var user string

					if err := rows.Scan(&user); err != nil {
						glog.Error(err)
					}
					ret = append(ret, user)
				}
			}
		}
	}
	return
}
