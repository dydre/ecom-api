-- name: ListProducts :many
SELECT 
    * 
FROM 
    productes;

-- name: FindProductsByID :one
SELECT * FROM productes WHERE id = $ 1;

