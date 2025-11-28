-- Insert 3 categories
INSERT INTO categories (code, name) VALUES
('CAT000', 'Uncategorized'),
('CAT001', 'Clothing'),
('CAT002', 'Shoes'),
('CAT003', 'Accessories');

-- Insert category to each product

-- Clothing
UPDATE products
SET category_id = (
    -- Subquery: Find the ID of the category based on its unique name.
    SELECT id
    FROM categories
    WHERE name = 'Clothing'
) 
WHERE code = 'PROD001';

UPDATE products
SET category_id = (
    SELECT id
    FROM categories
    WHERE name = 'Clothing'
) 
WHERE code = 'PROD004';

UPDATE products
SET category_id = (
    SELECT id
    FROM categories
    WHERE name = 'Clothing'
) 
WHERE code = 'PROD007';

-- Shoes
UPDATE products
SET category_id = (
    SELECT id
    FROM categories
    WHERE name = 'Shoes'
) 
WHERE code = 'PROD002';

UPDATE products
SET category_id = (
    SELECT id
    FROM categories
    WHERE name = 'Shoes'
)  
WHERE code = 'PROD006';

--Accessories
UPDATE products
SET category_id = (
    SELECT id
    FROM categories
    WHERE name = 'Accessories'
)
WHERE code = 'PROD003';

UPDATE products
SET category_id = (
    SELECT id
    FROM categories
    WHERE name = 'Accessories'
) 
WHERE code = 'PROD005';

UPDATE products
SET category_id = (
    SELECT id
    FROM categories
    WHERE name = 'Accessories'
)
WHERE code = 'PROD008';

-- Adding foreign key constraint
-- Add the NOT NULL constraint
ALTER TABLE products 
ALTER COLUMN category_id SET NOT NULL; 

-- Add the Foreign Key constraint
ALTER TABLE products 
ADD CONSTRAINT fk_products_category
FOREIGN KEY (category_id) 
REFERENCES categories(id)
    ON DELETE RESTRICT 
    ON UPDATE CASCADE;
    
-- setting default value for category in product table 
CREATE OR REPLACE FUNCTION get_uncategorized_id()
RETURNS integer AS
$$
    SELECT id
    FROM categories
    WHERE name = 'Uncategorized';
$$
LANGUAGE sql
IMMUTABLE; -- Use IMMUTABLE if you know the ID will not change once the row is created

-- Step 2: Set the default to call this function
ALTER TABLE products
ALTER COLUMN category_id SET DEFAULT get_uncategorized_id();

-- adding foreign key constraint to product table 
-- ALTER TABLE products 
--     ADD CONSTRAINT fk_products_category
--     FOREIGN KEY (category_id) 
--     REFERENCES categories(id)
--     ON DELETE RESTRICT 
--     ON UPDATE CASCADE;