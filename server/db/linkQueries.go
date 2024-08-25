package db

import (
	"database/sql"
	"mo_links/common"
)

func initializeLinkQueries() {
	incrementViewCountOfLinkStmt()
	addLinkStmt()
	matchingLinksStmt()
	getUserMoLinksStmt()
}

func incrementViewCountOfLinkStmt() *sql.Stmt {
	return getQuery(`
    UPDATE mo_links_entries SET views = views + 1
    WHERE organization_id = ? AND name = ?`)
}
func DbIncrementViewCountOfLink(organizationId int64, name string) error {
	_, err := incrementViewCountOfLinkStmt().Exec(organizationId, name)
	return err
}

func addLinkStmt() *sql.Stmt {
	return getQuery(`
	INSERT INTO mo_links_entries (url, name, created_by_user_id, organization_id) VALUES (?, ?, ?, ?)`)
}
func DbAddLink(url string, name string, userId int64, activeOrganizationId int64) error {
	_, err := addLinkStmt().Exec(url, name, userId, activeOrganizationId)
	if err != nil {
		return err
	}
	return nil
}

func getLinkStmt() *sql.Stmt {
	return getQuery(`
	SELECT id, name, url, organization_id, created_by_user_id FROM mo_links_entries WHERE id = ?`)
}
func DbGetLink(id int64) (common.MoLink, error) {
	var link common.MoLink
	err := getLinkStmt().QueryRow(id).Scan(&link.Id, &link.Name, &link.Url, &link.OrganizationId, &link.CreatedByUserId)
	return link, err
}

func matchingLinksStmt() *sql.Stmt {
	return getQuery(`
	SELECT url, organization_id FROM mo_links_entries
	WHERE organization_id = ?
	 AND name = ? ORDER BY created_at DESC`)
}
func DbGetMatchingLinks(organizationId int64, name string) ([]string, error) {
	rows, err := matchingLinksStmt().Query(organizationId, name)
	if err != nil {
		return []string{}, err
	}
	defer rows.Close()
	var links []string
	for rows.Next() {
		var url string
		var organizationId int64
		rows.Scan(&url, &organizationId)
		links = append(links, url)
	}
	return links, nil
}

func removeLinkStmt() *sql.Stmt {
	return getQuery(`
	DELETE FROM mo_links_entries WHERE id = ?`)
}
func DbRemoveLink(id int64) error {
	_, err := removeLinkStmt().Exec(id)
	return err
}

func updateLinkStmt() *sql.Stmt {
	return getQuery(`
	UPDATE mo_links_entries SET url = ? WHERE id = ?`)
}
func DbUpdateLink(id int64, url string) error {
	_, err := updateLinkStmt().Exec(url, id)
	return err
}

func getUserMoLinksStmt() *sql.Stmt {
	return getQuery(`
    SELECT id, name, url, organization_id, created_at, views FROM mo_links_entries WHERE created_by_user_id = ? AND organization_id = ?`)
}
func DbGetUserMoLinks(userId int64, organizationId int64) ([]common.MoLink, error) {
	rows, err := getUserMoLinksStmt().Query(userId, organizationId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var moLinks []common.MoLink
	for rows.Next() {
		var moLink common.MoLink
		err = rows.Scan(&moLink.Id, &moLink.Name, &moLink.Url, &moLink.OrganizationId, &moLink.CreatedAt, &moLink.Views)
		if err != nil {
			return nil, err
		}
		moLinks = append(moLinks, moLink)
	}
	return moLinks, nil
}
