
CREATE USER IF NOT EXISTS gdp IDENTIFIED WITH plaintext_password BY 'default';

GRANT ALL ON gdp.* TO gdp;