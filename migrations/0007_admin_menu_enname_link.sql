ALTER TABLE admin_menu
  ADD COLUMN IF NOT EXISTS enname varchar(128) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS iframe varchar(512) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS external varchar(512) NOT NULL DEFAULT '';

UPDATE admin_menu
SET enname = COALESCE(NULLIF(enname, ''), name)
WHERE COALESCE(enname, '') = '';
