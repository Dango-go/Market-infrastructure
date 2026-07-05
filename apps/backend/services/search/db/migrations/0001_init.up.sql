CREATE TABLE IF NOT EXISTS search_documents (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL,
    slug VARCHAR(180) NOT NULL,
    sku VARCHAR(120) NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    short_description TEXT NOT NULL DEFAULT '',
    category_slug VARCHAR(180) NOT NULL DEFAULT '',
    brand_slug VARCHAR(180) NOT NULL DEFAULT '',
    tags_csv TEXT NOT NULL DEFAULT '',
    specs_text TEXT NOT NULL DEFAULT '',
    price_cents BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(8) NOT NULL,
    available BOOLEAN NOT NULL DEFAULT FALSE,
    search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', COALESCE(name, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(sku, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(short_description, '')), 'B') ||
        setweight(to_tsvector('simple', COALESCE(tags_csv, '')), 'B') ||
        setweight(to_tsvector('simple', COALESCE(specs_text, '')), 'C') ||
        setweight(to_tsvector('simple', COALESCE(category_slug, '')), 'C') ||
        setweight(to_tsvector('simple', COALESCE(brand_slug, '')), 'C')
    ) STORED,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS search_documents_product_id_idx ON search_documents (product_id);
CREATE UNIQUE INDEX IF NOT EXISTS search_documents_slug_idx ON search_documents (slug);
CREATE INDEX IF NOT EXISTS search_documents_category_slug_idx ON search_documents (category_slug);
CREATE INDEX IF NOT EXISTS search_documents_brand_slug_idx ON search_documents (brand_slug);
CREATE INDEX IF NOT EXISTS search_documents_available_idx ON search_documents (available);
CREATE INDEX IF NOT EXISTS search_documents_search_vector_idx ON search_documents USING GIN (search_vector);
