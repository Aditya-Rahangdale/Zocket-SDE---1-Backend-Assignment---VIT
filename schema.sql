
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    product_description TEXT NOT NULL,
    product_images TEXT[] NOT NULL,
    compressed_product_images TEXT[],
    product_price NUMERIC(10, 2) NOT NULL
);
