CREATE TYPE product_type AS ENUM ('8b0bf29c-58e8-4310-8bb1-a1b9771f9c47','2b98f424-91c9-46cc-abd7-c888208807da', 'a19a514e-41c9-4666-a01a-e3f9c0255609');
CREATE TYPE custom_field_type AS ENUM ('8b0bf29c-58e8-4310-8bb1-a1b9771f9c47','2b98f424-91c9-46cc-abd7-c888208807da', 'a19a514e-41c9-4666-a01a-e3f9c0255609');

CREATE TABLE "measurement_precision" (
    "id" UUID PRIMARY KEY,
    "value" VARCHAR NOT NULL
);



CREATE TYPE user_type AS ENUM (
    '1fe92aa8-2a61-4bf1-b907-182b497584ad', -- system user
    '9fb3ada6-a73b-4b81-9295-5c1605e54552'  -- admin user
);

CREATE TYPE app_type AS ENUM (
    '1fe92aa8-2a61-4bf1-b907-182b497584ad', -- client
    '9fb3ada6-a73b-4b81-9295-5c1605e54552'  -- admin
);

CREATE TABLE IF NOT EXISTS "user" (
    "id" UUID PRIMARY KEY,
    "user_type_id" user_type NOT NULL,
    "first_name" VARCHAR(250) NOT NULL,
    "last_name" VARCHAR(250) NOT NULL,
    "phone_number" VARCHAR(30) NOT NULL,
    "image" TEXT,
    "deleted_at" BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX "user_deleted_at_idx" ON "user"("deleted_at");

INSERT INTO "user" (
    "id",
    "first_name",
    "last_name",
    "phone_number",
    "user_type_id"
) VALUES (
    '9a2aa8fe-806e-44d7-8c9d-575fa67ebefd',
    'admin',
    'admin',
    '99894172774',
    '9fb3ada6-a73b-4b81-9295-5c1605e54552'
);

CREATE TABLE IF NOT EXISTS "default_measurement_unit" (
    "id" UUID PRIMARY KEY,
    "long_name" VARCHAR NOT NULL,
    "short_name" VARCHAR NOT NULL,
    "long_name_translation" JSONB NOT NULL,
    "short_name_translation" JSONB NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID REFERENCES "user"("id") ON DELETE SET NULL,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "deleted_by" UUID REFERENCES "user"("id") ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS "company" (
    "id" UUID PRIMARY KEY,
    "name" VARCHAR(64) NOT NULL,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "created_by" UUID REFERENCES "user"("id") ON DELETE SET NULL
);
CREATE INDEX company_deleted_at_idx ON "company"("deleted_at");

CREATE TABLE IF NOT EXISTS "company_user" (
    "user_id" UUID NOT NULL REFERENCES "user" ("id"),
    "company_id" UUID NOT NULL,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    PRIMARY KEY("user_id", "company_id", "deleted_at")
);

CREATE INDEX "company_user_deleted_at_idx" ON "company_user"("deleted_at");


CREATE TABLE IF NOT EXISTS "shop" (
    "id" UUID PRIMARY KEY,
    "name" VARCHAR(64) NOT NULL,
    "company_id" UUID NOT NULL,
    "created_by" UUID,
    "deleted_at" BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX shop_deleted_at_idx ON "shop"("deleted_at");


CREATE TABLE IF NOT EXISTS "brand" (
    "id" UUID PRIMARY KEY,
    "name" VARCHAR(200) NOT NULL,
    "company_id" UUID NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID REFERENCES "user"("id") ON DELETE SET NULL,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "deleted_by" UUID REFERENCES "user"("id") ON DELETE SET NULL,
    UNIQUE ("name", "company_id", "deleted_at")
);
CREATE INDEX brand_deleted_at_idx ON "brand" ("deleted_at");

CREATE TABLE IF NOT EXISTS "measurement_unit" (
    "id" UUID PRIMARY KEY,
    "company_id" UUID,
    "is_deletable" BOOLEAN NOT NULL DEFAULT TRUE,
    "unit_id" UUID NOT NULL REFERENCES "default_measurement_unit"("id") ON DELETE CASCADE,
    "precision_id" UUID NOT NULL REFERENCES "measurement_precision"("id") ,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID  REFERENCES "user"("id") ON DELETE SET NULL,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "deleted_by" UUID REFERENCES "user"("id") ON DELETE SET NULL,
    UNIQUE ("unit_id", "precision_id", "company_id", "deleted_at")
);
CREATE INDEX measurement_unit_deleted_at_idx ON "measurement_unit"("deleted_at");

CREATE TABLE IF NOT EXISTS "supplier" (
    "id" UUID PRIMARY KEY,
    "name" VARCHAR NOT NULL,
    "company_id" UUID NOT NULL,
    "created_by" UUID,
    "deleted_at" BIGINT NOT NULL DEFAULT 0
);
CREATE INDEX supplier_deleted_at_idx ON "supplier"("deleted_at");

CREATE TABLE IF NOT EXISTS "category" (
    "id" UUID PRIMARY KEY,
    "name" VARCHAR(100) NOT NULL,
    "parent_id" UUID REFERENCES "category"("id") ON DELETE SET NULL,
    "company_id" UUID NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID  REFERENCES "user"("id") ON DELETE SET NULL,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "deleted_by" UUID REFERENCES "user"("id") ON DELETE SET NULL,
    UNIQUE ("name", "company_id", "deleted_at")
);
CREATE INDEX category_deleted_at_idx ON "category"("deleted_at");


CREATE TABLE IF NOT EXISTS "product" (
    "id" UUID PRIMARY KEY,
    "company_id" UUID NOT NULL,
    "product_type_id" product_type NOT NULL,
    "parent_id" UUID REFERENCES "product"("id") ON DELETE SET NULL,
    "last_version" int NOT NULL DEFAULT 1,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "deleted_by" UUID
);

CREATE INDEX product_deleted_at_idx ON "product" ("deleted_at");


CREATE TABLE "product_detail" (
    "id" UUID PRIMARY KEY,
    "sku" VARCHAR NOT NULL,
    "product_id"  UUID NOT NULL REFERENCES "product"("id") ON DELETE CASCADE,
    "version" INT NOT NULL DEFAULT 1,
    "name" TEXT NOT NULL,
    "mxik_code" VARCHAR,
    "is_marking" BOOLEAN NOT NULL DEFAULT FALSE,
    "brand_id" UUID REFERENCES "brand"("id"),
    "description" TEXT,
    "measurement_unit_id" UUID REFERENCES "measurement_unit"("id")  ON DELETE SET NULL,
    "supplier_id" UUID REFERENCES "supplier"("id") ON DELETE SET NULL,
    "vat_id" UUID REFERENCES "vat"("id") ON DELETE SET NULL,
    "created_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID,
    UNIQUE("product_id", "version")
);

CREATE TABLE "product_image" (
    "id" UUID PRIMARY KEY,
    "sequence_number"  INT NOT NULL DEFAULT 0,
    "product_detail_id" UUID NOT NULL REFERENCES "product_detail"("id") ON DELETE CASCADE,
    "file_name" TEXT NOT NULL,
    UNIQUE("product_detail_id", "sequence_number")
);


CREATE TABLE IF NOT EXISTS "product_barcode" (
    "barcode" VARCHAR(300) NOT NULL,
    "product_detail_id" UUID NOT NULL REFERENCES "product_detail"("id") ON DELETE CASCADE,
    PRIMARY KEY ("barcode", "product_detail_id")
);

CREATE TABLE IF NOT EXISTS "product_category" (
    "product_detail_id" UUID REFERENCES "product_detail"("id") ON DELETE CASCADE,
    "category_id" UUID REFERENCES "category"("id") ON DELETE CASCADE,
    PRIMARY KEY ("product_detail_id", "category_id")
);

CREATE TABLE "tag" (
    "id" UUID PRIMARY KEY,
    "company_id" UUID NOT NULL,
    "name" VARCHAR(50) NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID ON DELETE SET NULL,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "deleted_by" UUIDON DELETE SET NULL,
    UNIQUE ("company_id", "name", "deleted_at")
);
CREATE INDEX tag_deleted_at ON "tag" ("deleted_at");

CREATE TABLE IF NOT EXISTS "product_tag" (
    "product_detail_id" UUID NOT NULL REFERENCES "product_detail"("id") ON DELETE CASCADE,
    "tag_id" UUID NOT NULL REFERENCES "tag"("id") ON DELETE CASCADE,
    PRIMARY KEY("product_detail_id", "tag_id")
);


CREATE TABLE IF NOT EXISTS "measurement_values" (
    "shop_id" UUID NOT NULL,
    "amount" NUMERIC NOT NULL DEFAULT 0,
    "total" NUMERIC NOT NULL DEFAULT 0,
    "has_trigger" BOOLEAN NOT NULL DEFAULT FALSE,
    "small_left" NUMERIC NOT NULL DEFAULT 0,
    "total_imported" NUMERIC NOT NULL DEFAULT 0,
    "total_sold" NUMERIC NOT NULL DEFAULT 0,
    "total_transfered" NUMERIC NOT NULL DEFAULT 0,
    "total_transfer_arrived" NUMERIC NOT NULL DEFAULT 0,
    "total_supplier_order" NUMERIC NOT NULL DEFAULT 0,
    "total_postpone_order" NUMERIC NOT NULL DEFAULT 0,
    "is_available" BOOLEAN NOT NULL DEFAULT true,
    "product_id" UUID NOT NULL REFERENCES "product"("id") ON DELETE CASCADE,
    PRIMARY KEY("product_id", "shop_id")
);

alter table shop_price
    add constraint product_id_shop_id_pfkey
        primary key (product_id, shop_id);


CREATE TABLE IF NOT EXISTS "shop_price" (
    "id" UUID NOT NULL,
    "supply_price" NUMERIC NOT NULL DEFAULT 0,
    "min_price" NUMERIC NOT NULL DEFAULT 0,
    "max_price" NUMERIC NOT NULL DEFAULT 0,
    "retail_price" NUMERIC NOT NULL DEFAULT 0,
    "whole_sale_price" NUMERIC NOT NULL DEFAULT 0,
    "shop_id" UUID NOT NULL,
    "product_id" UUID NOT NULL REFERENCES "product"("id") ON DELETE CASCADE,
    PRIMARY KEY("product_id", "shop_id")
);

CREATE TABLE "custom_field" (
    "id" UUID PRIMARY KEY,
    "company_id" UUID NOT NULL,
    "name" VARCHAR(100) NOT NULL,
    "type" custom_field_type NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID ON DELETE SET NULL,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "deleted_by" UUID ON DELETE SET NULL,
    UNIQUE ("company_id", "name", "deleted_at")
);
CREATE INDEX custom_field_deleted_at_idx ON "custom_field"("deleted_at");

CREATE TABLE "product_cf" (
    "product_detail_id" UUID REFERENCES "product_detail" ON DELETE CASCADE,
    "custom_field_id" UUID REFERENCES "custom_field"("id") ON DELETE CASCADE,
    "value" TEXT NOT NULL,
    UNIQUE ("product_detail_id", "custom_field_id")
);

CREATE TABLE IF NOT EXISTS "sku" (
    "company_id" UUID NOT NULL,
    "value" INT,
    UNIQUE ("company_id", "value")
);

CREATE TABLE IF NOT EXISTS "vat" (
    "id" UUID PRIMARY KEY,
    "name" VARCHAR NOT NULL,
    "percentage" NUMERIC NOT NULL DEFAULT 0,
    "company_id" UUID NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "deleted_by" UUID
);
CREATE INDEX vat_deleted_at_idx ON "vat"("deleted_at");


CREATE TABLE IF NOT EXISTS "label" (
    "id" UUID PRIMARY KEY,
    "company_id" UUID NOT NULL,
    "name" VARCHAR NOT NULL,
    "width" INT NOT NULL,
    "height" INT NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "deleted_by" UUID,
    UNIQUE ("company_id", "name")
);
CREATE INDEX label_deleted_at_idx ON "label"("deleted_at");


CREATE TABLE IF NOT EXISTS "label_content" (
    "id" UUID PRIMARY KEY,
    "label_id" UUID NOT NULL REFERENCES "label"("id") ON DELETE CASCADE,
    "position_x" INT NOT NULL DEFAULT 0,
    "position_y" INT NOT NULL DEFAULT 0,
    "width" INT NOT NULL,
    "height" INT NOT NULL,
    "type" VARCHAR NOT NULL,
    "product_image" VARCHAR NOT NULL DEFAULT '',
    "field_name" VARCHAR NOT NULL,
    "font_family" VARCHAR NOT NULL,
    "font_style" VARCHAR NOT NULL DEFAULT 'normal',
    "font_size" INT NOT NULL,
    "font_weight" INT NOT NULL,
    "text_align" VARCHAR NOT NULL DEFAULT 'center',
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "deleted_by" UUID
);
CREATE INDEX label_content_deleted_at_idx ON "label_content"("deleted_at");

CREATE OR REPLACE FUNCTION create_defaults()
  RETURNS TRIGGER
  LANGUAGE PLPGSQL
  AS
$$
BEGIN
    INSERT INTO "measurement_unit" ("id", "company_id", "unit_id", "is_deletable", "precision_id", "created_by") VALUES
        (uuid_generate_v4(), NEW."id", '65412957-a736-4916-b7b3-4b6157967b0a', FALSE, 'abd0b9ae-1c5e-4cd0-96d2-768f3249d25c', NEW.created_by),--Kg
        (uuid_generate_v4(), NEW."id", '490b83e8-0e61-405a-a466-57923cedf2f5', FALSE, 'abd0b9ae-1c5e-4cd0-96d2-768f3249d25c', NEW.created_by);
    INSERT INTO "vat" ("id", "name", "percentage", "company_id") VALUES
        (uuid_generate_v4(), 'Standart', 12, NEW."id"),
        (uuid_generate_v4(), 'Reduced', 6, NEW."id"),
        (uuid_generate_v4(), 'No Vat', 0, NEW."id");
	RETURN NEW;
END;
$$;

CREATE TABLE "scales_template" (
    "id" UUID PRIMARY KEY,
    "product_unit_ids" VARCHAR,
    "company_id" UUID NOT NULL,
    "name" VARCHAR NOT NULL,
    "value" VARCHAR NOT NULL,
    "created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "created_by" UUID,
    "deleted_at" BIGINT NOT NULL DEFAULT 0,
    "deleted_by" UUID
);

CREATE OR REPLACE FUNCTION create_scale_template_defaults()
  RETURNS TRIGGER
  LANGUAGE PLPGSQL
  AS
$$
BEGIN
    IF NEW."unit_id" = '65412957-a736-4916-b7b3-4b6157967b0a' THEN -- kg
        INSERT INTO "scales_template" ("id", "company_id", "product_unit_ids", "name", "value", "created_by") VALUES
            (uuid_generate_v4(), NEW."company_id", NEW."id", 'Mettler toledo Spct 1', '{sku},{sku},0,{price},0,0,0,0,0,0,0,0,0,{name}', NEW.created_by),
            (uuid_generate_v4(), NEW."company_id", NEW."id", 'Shtrix_M', '{sku};{name};;{price};0;0;0;{sku};0;0;;01.01.01;0;0;0;0;01.01.01', NEW.created_by),
            (uuid_generate_v4(), NEW."company_id", NEW."id", 'RLS1200', '{name};{sku};{sku};7;{price};29', NEW.created_by);
    END IF;
	RETURN NEW;
END;
$$;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- triggers

CREATE OR REPLACE FUNCTION check_measurement_unit_is_deletable()
  RETURNS TRIGGER
  LANGUAGE PLPGSQL
  AS
$$
BEGIN
    IF NEW."deleted_at" <> 0 AND OLD."is_deletable" = FALSE THEN
        RAISE exception '%v is not deletable', NEW."id";
        NEW."deleted_at"=0;
    END IF;
    RETURN NEW;
END;
$$;


-- triggers
CREATE OR REPLACE TRIGGER check_measurement_unit_is_deletable
    BEFORE UPDATE ON "measurement_unit"
    FOR EACH ROW
    EXECUTE PROCEDURE check_measurement_unit_is_deletable();


CREATE OR REPLACE TRIGGER create_defaults
    AFTER INSERT ON "company"
    FOR EACH ROW
    EXECUTE PROCEDURE create_defaults();


CREATE OR REPLACE TRIGGER create_default_scales_templates
    BEFORE INSERT ON "measurement_unit"
    FOR EACH ROW
    EXECUTE PROCEDURE create_scale_template_defaults();

INSERT INTO measurement_precision("id", "value")
VALUES
('abd0b9ae-1c5e-4cd0-96d2-768f3249d25c', '1'),
('e2e9537a-8b74-42da-8f24-9e6cd53e06e5', '.0'),
('2bf0ff0c-b5cd-4bad-a5ff-556a6acab501', '.00'),
('fd5571b1-851c-4493-8bb8-9bb7eed0f09f', '.000'),
('8a2aa4ad-7370-44ca-8a0b-96072534aa73', '.0000'),
('e899c30d-d355-4398-9a79-c05ca76f28ed', '.00000');


INSERT INTO default_measurement_unit(id, long_name, short_name, long_name_translation, short_name_translation)
values
(uuid_generate_v4(), 'Ар (100 м2)',	'а', '{"ru":"Ар (100 м2)","uz":"Ar (100 m2)", "en":"Ar (100 m2)"}', '{"ru":"Ар (100 м2)","uz":"Ar (100 m2)", "en":"Ar (100 m2)"}'),
(uuid_generate_v4(), 'Гектар', 'га', '{"ru":"Гектар (100 м2)","uz":"Gektar (100 m2)", "en":"Hectare (100 m2)"}', '{"ru":"Га (100 м2)","uz":"Gek (100 m2)", "en":"Ha (100 m2)"}'),
(uuid_generate_v4(), 'Грамм', 'г', '{"ru":"Грамм","uz":"Gramm", "en":"Gram"}', '{"ru":"Гр","uz":"Gr", "en":"G"}'),
(uuid_generate_v4(), 'Дециметр', 'дм', '{"ru":"Дециметр","uz":"Detsimetr", "en":"Decimeter"}', '{"ru":"дм","uz":"dm", "en":"dm"}'),
(uuid_generate_v4(), 'Дюйм (25,4 мм)', 'дюйм', '{"ru":"Дюйм","uz":"Dyuym", "en":"Inch"}', '{"ru":"дм","uz":"dm", "en":"in"}'),
(uuid_generate_v4(), 'Квадратный дециметр',	'дм2', '{"ru":"Квадратный дециметр","uz":"Kvadrat detsimetr", "en":"Square decimeter"}', '{"ru":"дм2","uz":"dm2", "en":"dm2"}'),
(uuid_generate_v4(), 'Квадратный дюйм (645,16 мм2)',	'дюйм2', '{"ru":"Квадратный дюйм","uz":"Kvadrat dyuym", "en":"Square inch"}', '{"ru":"дюйм2","uz":"dyum2", "en":"in2"}'),
(uuid_generate_v4(), 'Квадратный километр',	'км2', '{"ru":"Квадратный километр","uz":"Kvadrat km", "en":"Square kilometеr"}', '{"ru":"км2","uz":"km2", "en":"km2"}'),
(uuid_generate_v4(), 'Квадратный метр',	'м2', '{"ru":"Квадратный метр","uz":"Kvadrat metr", "en":"Square metr"}', '{"ru":"м2","uz":"m2", "en":"m2"}'),
(uuid_generate_v4(), 'Квадратный миллиметр',	'мм2', '{"ru":"Квадратный миллиметр","uz":"Kvadrat millimetr", "en":"Square millimeter"}', '{"ru":"мм2","uz":"mm2", "en":"mm2"}'),
(uuid_generate_v4(), 'Квадратный сантиметр',	'см2', '{"ru":"Квадратный сантиметр","uz":"Kvadrat santimetr", "en":"Square centimeter"}', '{"ru":"см2","uz":"sm2", "en":"cm2"}'),
(uuid_generate_v4(), 'Квадратный фут (0,092903 м2)',	'фут2', '{"ru":"Квадратный фут","uz":"Kvadrat fut", "en":"Square foot"}', '{"ru":"фут2","uz":"ft2", "en":"ft2"}'),
('65412957-a736-4916-b7b3-4b6157967b0a', 'Килограмм',	'кг', '{"ru":"Килограмм","uz":"Kilogram", "en":"Kilogramme"}', '{"ru":"кг","uz":"kg", "en":"kg"}'),
(uuid_generate_v4(), 'Километр',	'км', '{"ru":"Километр","uz":"Kilometr", "en":"Kilometer"}', '{"ru":"км","uz":"km", "en":"km"}'),
(uuid_generate_v4(), 'Кубический дюйм (16387,1 мм3)',	'дюйм3', '{"ru":"Кубический дюйм","uz":"Dyuym", "en":"Inch"}', '{"ru":"дюйм3","uz":"dyum3", "en":"in3"}'),
(uuid_generate_v4(), 'Кубический метр',	'м3', '{"ru":"Кубический метр","uz":"Kub metr", "en":"Cubic meter"}', '{"ru":"м3","uz":"m3", "en":"m3"}'),
(uuid_generate_v4(), 'Кубический миллиметр',	'мм3', '{"ru":"Кубический миллиметр","uz":"Kub millimetr", "en":"Cubic millimeter"}', '{"ru":"мм3","uz":"mm3", "en":"mm3"}'),
(uuid_generate_v4(), 'Кубический сантиметр', 	'см3', '{"ru":"Кубический сантиметр","uz":"Kub santimetr", "en":"Cubic centimeter"}', '{"ru":"см3","uz":"sm3", "en":"sm3"}'),
(uuid_generate_v4(), 'Милллилитр',	'мл', '{"ru":"Милллилитр","uz":"Millilitr", "en":"Milliliter"}', '{"ru":"мл","uz":"ml", "en":"ml"}'),
(uuid_generate_v4(), 'Кубический фут (0,02831685 м3)',	'фут3', '{"ru":"Кубический фут","uz":"Kub dyuym", "en":"Cubic inch"}', '{"ru":"фут3","uz":"ft3", "en":"in3"}'),
(uuid_generate_v4(), 'Литр',	'л', '{"ru":"Литр","uz":"Litr", "en":"Liter"}', '{"ru":"л","uz":"l", "en":"l"}'),
(uuid_generate_v4(), 'Кубический дециметр',	'дм3', '{"ru":"Кубический дециметр","uz":"Kub detsimetr", "en":"Cubic decimeter"}', '{"ru":"дм3","uz":"dm3", "en":"dm3"}'),
(uuid_generate_v4(), 'Месяц',	'мес', '{"ru":"Месяц","uz":"Oy", "en":"Month"}', '{"ru":"мес","uz":"oy", "en":"mon"}'),
(uuid_generate_v4(), 'Метр',	'м', '{"ru":"Метр","uz":"Metr", "en":"Meter"}', '{"ru":"м","uz":"m", "en":"m"}'),
(uuid_generate_v4(), 'Метрический карат',	'кар', '{"ru":"Метрический карат","uz":"Metrik karat", "en":"Metric carat"}', '{"ru":"кар","uz":"mk", "en":"mc"}'),
(uuid_generate_v4(), 'Миллиграмм',	'мг', '{"ru":"Миллиграмм","uz":"Milligram", "en":"Milligram"}', '{"ru":"мг","uz":"mg", "en":"mg"}'),
(uuid_generate_v4(), 'Миллиметр',	'мм', '{"ru":"Миллиметр","uz":"Millimetr", "en":"Millemeter"}', '{"ru":"мм","uz":"mm", "en":"mm"}'),
(uuid_generate_v4(), 'Минута',	'мин', '{"ru":"Минута","uz":"Minut", "en":"Minute"}', '{"ru":"мин","uz":"min", "en":"min"}'),
(uuid_generate_v4(), 'Рулон',	'рулон', '{"ru":"Рулон","uz":"Rulon", "en":"Roll"}', '{"ru":"рулон","uz":"rulon", "en":"roll"}'),
(uuid_generate_v4(), 'Сантиметр',	'см', '{"ru":"Сантиметр","uz":"Santimetr", "en":"Santimeter"}', '{"ru":"см","uz":"sm", "en":"sm"}'),
(uuid_generate_v4(), 'Секунда',	'с', '{"ru":"Секунда","uz":"Sekund", "en":"Second"}', '{"ru":"с","uz":"sek", "en":"sec"}'),
(uuid_generate_v4(), 'Сутки',	'сут', '{"ru":"Сутки","uz":"Sutka", "en":"Day"}', '{"ru":"сут","uz":"sut", "en":"day"}'),
(uuid_generate_v4(), 'Тонна',	'т', '{"ru":"Тонна","uz":"Tonna", "en":"Ton"}', '{"ru":"т","uz":"t", "en":"t"}'),
(uuid_generate_v4(), 'Фут (0,3048 м)',	'фут', '{"ru":"Фут","uz":"Fut", "en":"Foot"}', '{"ru":"фут","uz":"ft", "en":"ft"}'),
(uuid_generate_v4(), 'Центнер',	'ц', '{"ru":"Центнер","uz":"Sentner", "en":"Centner"}', '{"ru":"ц","uz":"s", "en":"c"}'),
(uuid_generate_v4(), 'Час',	'ч', '{"ru":"Час","uz":"Soat", "en":"Hour"}', '{"ru":"ч","uz":"s", "en":"h"}'),
('490b83e8-0e61-405a-a466-57923cedf2f5', 'Штука',	'шт', '{"ru":"Штука","uz":"Dona", "en":"Piece"}', '{"ru":"шт","uz":"d", "en":"pc"}');
