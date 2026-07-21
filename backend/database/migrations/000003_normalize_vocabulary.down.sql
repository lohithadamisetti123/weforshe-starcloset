-- Down: restore the wide constraint (we can't recover which specific style
-- each product had before normalisation, so we just allow all values again).
BEGIN;
ALTER TABLE products DROP CONSTRAINT IF EXISTS chk_products_style;
ALTER TABLE products
    ADD CONSTRAINT chk_products_style CHECK (style IN (
        'Ethnic','Western','Formal','Streetwear',
        'Minimal','Bohemian','Athleisure','Indo-Western',
        'Traditional','Contemporary Ethnic','Luxury Ethnic',
        'Festive','Elegant','Casual',
        'Smart Casual','Business Casual',
        'Oversized','Slim Fit','Regular Fit','Relaxed Fit',
        'Floral','Printed','Graphic','Embroidered',
        'Chikankari','Bandhani','Mirror Work','Gota Patti',
        'Jacquard','Lucknowi','Organza','Sharara','Gharara','Angrakha',
        'Chanderi','Georgette','Chiffon','Velvet','Silk',
        'Cotton','Linen','Denim','Leather','Suede',
        'Vintage','Classic','Modern','Luxury','Premium',
        'Minimalist','Resort','Lounge','Sporty','Boho','Contemporary',
        'Traditional Decor','Industrial','Natural','Matte','Glossy','Scandinavian'
    ));
COMMIT;
