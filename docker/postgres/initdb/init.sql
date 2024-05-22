-- Подключение к базе данных postgres для создания новой базы данных
\connect postgres

-- Создание базы данных, если она не существует
DO
$do$
BEGIN
   IF NOT EXISTS (
      SELECT FROM pg_database
      WHERE datname = 'app_update_db') THEN
      CREATE DATABASE app_update_db;
   END IF;
END
$do$;

-- Подключение к созданной базе данных
\connect app_update_db

-- Создание таблицы для хранения данных об обновлениях приложений
CREATE TABLE IF NOT EXISTS app_updates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    package_name VARCHAR(255) NOT NULL,
    downloads JSONB NOT NULL
);

-- Вставка тестовых данных
INSERT INTO app_updates (name, package_name, downloads) VALUES (
    'Example App',
    'com.example.app',
    '{
        "googleplay": {
            "url": "https://example.com/download/app",
            "latestVersion": {
                "versionCode": 20,
                "versionName": "20.0.0",
                "checksum": "d41d8cd98f00b204e9800998ecf8427e",
                "fileSize": 12345678,
                "updateDescription": "Major update with new features and improvements."
            },
            "updateRequired": {
                "hardUpdate": {
                    "minimumVersionCode": 1
                },
                "softUpdate": {
                    "minimumVersionCode": 2
                }
            }
        },
        "appgallery": {
            "url": "https://example.com/download/app_gallery",
            "latestVersion": {
                "versionCode": 11,
                "versionName": "11.0.0",
                "checksum": "d41d8cd98f00b204e9800998ecf8427e",
                "fileSize": 9876543,
                "updateDescription": "Initial release."
            },
            "updateRequired": {
                "hardUpdate": {
                    "minimumVersionCode": 0
                },
                "softUpdate": {
                    "minimumVersionCode": 1
                }
            }
        }
    }'
);
