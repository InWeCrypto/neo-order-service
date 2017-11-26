DROP INDEX IF EXISTS NEO_ORDER_TX;
DROP TABLE IF EXISTS NEO_ORDER;

DROP INDEX IF EXISTS NEO_WALLET_ADDRESS_USER;
DROP TABLE IF EXISTS NEO_WALLET;


CREATE TABLE NEO_ORDER (
  "id"          SERIAL PRIMARY KEY,
  "tx"          VARCHAR(128) NOT NULL, -- tx by which the utxo createdg'g
  "from"        VARCHAR(128) NOT NULL, -- value out address
  "to"          VARCHAR(128) NOT NULL, -- value out address
  "asset"       VARCHAR(128) NOT NULL, -- asset type string
  "value"       NUMERIC      NOT NULL, -- utxo value
  "createTime"  TIMESTAMP    NOT NULL DEFAULT NOW(), -- utxo create time
  "confirmTime" TIMESTAMP  -- order confirmTime
);

CREATE INDEX NEO_ORDER_TX
  ON NEO_ORDER ("tx", "from", "to", "asset");

CREATE TABLE NEO_WALLET (
  "id"         SERIAL PRIMARY KEY,
  "address"    VARCHAR(128) NOT NULL,
  "userid"     VARCHAR(128) NOT NULL, -- user id for message push
  "createTime" TIMESTAMP    NOT NULL
);


CREATE UNIQUE INDEX NEO_WALLET_ADDRESS_USER
  ON NEO_WALLET ("address", "userid");