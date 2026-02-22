package repository

import(
	"database/sql"
	"github.com/sachinchandra/goflash/internal/database"
	"log"
)

func IntializeSchema(){
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name TEXT,
		quantity INT
	);
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		product_id INT,
		user_id INT
	);
	`

	_,err:=database.DB.Exec(createTableSQL)
	if err !=nil{
		log.Fatal("Failed to create tables: ", err)
	}

	database.DB.Exec("DELETE FROM orders")
	database.DB.Exec("DELETE FROM products")
	database.DB.Exec("INSERT INTO products(id,name,quantity) VALUES (1,'iPhone 17', 100)")
}


func PurchaseProductNaive(productID int, userID int) error{
	var quantity int 
	err := database.DB.QueryRow("SELECT quantity FROM products WHERE id= $1", productID).Scan(&quantity)
	if err !=nil{
		return err
	}
	if quantity<=0{
		return sql.ErrNoRows
	}

	_,err =database.DB.Exec("UPDATE products SET quantity = quantity-1 WHERE id=$1", productID)
	if err !=nil{
		return err
	}
	_, err=database.DB.Exec("INSERT INTO orders(product_id, user_id) VALUES ($1,$2)", productID,userID)
	return err
}

func PurchaseProductPessimistic(productID int, userID int) error{
	tx, err := database.DB.Begin()
	if err !=nil{
		return err
	}

	defer tx.Rollback()
	var quantity int
	err = tx.QueryRow("SELECT quantity FROM products WHERE id = $1 FOR UPDATE", productID).Scan(&quantity)

	if err !=nil{
		return err
	}
	if quantity <=0{
		return sql.ErrNoRows
	}

	_, err = tx.Exec("UPDATE products SET quantity = quantity - 1 WHERE id = $1", productID)
	if err !=nil{
		return err
	}

	_, err = tx.Exec("INSERT INTO orders (product_id, user_id) VALUES ($1, $2)", productID, userID)
	if err != nil {
		return err
	}

	return tx.Commit()

}

