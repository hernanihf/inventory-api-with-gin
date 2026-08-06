INSERT INTO products (name, price, stock)
VALUES ('coca', 2.0, 100),
       ('sprite', 1.5, 50),
       ('lays', 3.33, 2),
       ('doritos', 2.45, 42),
       ('oreos', 1.75, 0),
       ('toblerone', 5.66, 33) ON CONFLICT (name) DO NOTHING;