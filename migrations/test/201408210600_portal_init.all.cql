CREATE TABLE IF NOT EXISTS user (
  id uuid,
  name text,
  email text,
  teams map<uuid,text>,
  PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS user_email (
  email text,
  user uuid,
  PRIMARY KEY (email)
);

CREATE TABLE IF NOT EXISTS team (
  id uuid,
  name text,
  members map<uuid,text>,
  bubbles set<uuid>,
  ips     map<inet,text>,
  PRIMARY KEY (id)
);
CREATE TABLE IF NOT EXISTS network (
  id uuid,
  bubble uuid,
  region text,
  networks map<text,text>,
  PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS network_region ON network(region);
