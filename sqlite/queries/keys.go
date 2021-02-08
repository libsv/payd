package queries

const (
	KeyByName = `
	SELECT name, xpriv, createdAt
	FROM keys
	WHERE name 
	`

	CreateKey = `
	INSERT INTO keys(name, xpriv, createdat)
	VALUES(:name, :xpriv,:createdat)
	`
)
