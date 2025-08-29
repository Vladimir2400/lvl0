-- Таблица для хранения информации о заказах
CREATE TABLE orders (
    order_uid VARCHAR(255) PRIMARY KEY,
    track_number VARCHAR(255),
    entry VARCHAR(255),
    locale VARCHAR(10),
    internal_signature VARCHAR(255),
    customer_id VARCHAR(255),
    delivery_service VARCHAR(255),
    shardkey VARCHAR(10),
    sm_id INT,
    date_created TIMESTAMP WITH TIME ZONE,
    oof_shard VARCHAR(10)
);
    
-- Таблица для информации о доставке, связанная с заказом
CREATE TABLE deliveries (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
    name VARCHAR(255),
    phone VARCHAR(255),
    zip VARCHAR(255),
    city VARCHAR(255),
    address VARCHAR(255),
    region VARCHAR(255),
    email VARCHAR(255)
);
   
-- Таблица для информации об оплате, связанная с заказом
CREATE TABLE payments (
     id SERIAL PRIMARY KEY,
     order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
     transaction VARCHAR(255),
     request_id VARCHAR(255),
     currency VARCHAR(10),
     provider VARCHAR(255),
     amount INT,
     payment_dt BIGINT,
     bank VARCHAR(255),
     delivery_cost INT,
     goods_total INT,
     custom_fee INT
);
   
-- Таблица для товаров в заказе, связанная с заказом
CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    order_uid VARCHAR(255) REFERENCES orders(order_uid) ON DELETE CASCADE,
    chrt_id INT,
    track_number VARCHAR(255),
    price INT,
    rid VARCHAR(255),
    name VARCHAR(255),
    sale INT,
    size VARCHAR(255),
    total_price INT,
    nm_id INT,
    brand VARCHAR(255),
    status INT
);