-- 사이트 정보 테이블
CREATE TABLE IF NOT EXISTS sites (
    site_id INT AUTO_INCREMENT PRIMARY KEY,
    code VARCHAR(50) NOT NULL,             -- 사이트 구분 코드 (예: 'cloud', 'store')
    name VARCHAR(255) NOT NULL,            -- 사이트 이름
    domain VARCHAR(255) NOT NULL DEFAULT '',                   -- 도메인 (옵션)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT NULL,
    UNIQUE KEY (code)
);

-- 페이지 그룹 테이블
CREATE TABLE IF NOT EXISTS page_groups (
    group_id INT AUTO_INCREMENT PRIMARY KEY,
    site_id INT NOT NULL,
    name VARCHAR(255) NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (site_id) REFERENCES sites(site_id) ON DELETE CASCADE,
    UNIQUE KEY (site_id, name)
);

-- 페이지와 메뉴 관리를 위한 테이블
CREATE TABLE IF NOT EXISTS pages (
    page_id INT AUTO_INCREMENT PRIMARY KEY,
    site_id INT NOT NULL,
    group_id INT NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    slug VARCHAR(255) NOT NULL DEFAULT '',
    parent_id INT NULL DEFAULT NULL,
    depth INT NOT NULL DEFAULT 0,
    content TEXT NOT NULL DEFAULT '',
    menu_order INT NOT NULL DEFAULT 0,
    is_published BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT NULL,
    FOREIGN KEY (parent_id) REFERENCES pages(page_id) ON DELETE CASCADE,
    FOREIGN KEY (site_id) REFERENCES sites(site_id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES page_groups(group_id) ON DELETE CASCADE,
    UNIQUE KEY (site_id, slug, parent_id),
    INDEX idx_group_id (group_id)
);

DELIMITER //
CREATE TRIGGER before_pages_update
BEFORE UPDATE ON pages
FOR EACH ROW
BEGIN
    IF NEW.is_published != OLD.is_published THEN
        SET NEW.updated_at = CURRENT_TIMESTAMP;
    END IF;
END//
DELIMITER ;

-- 샘플 사이트 데이터
INSERT INTO sites (name, code) VALUES 
('클라우드', 'cloud'),
('스토어', 'store');

SET @cloud_site_id = (SELECT site_id FROM sites WHERE code = 'cloud');

-- 클라우드 사이트 그룹 샘플 데이터
INSERT INTO page_groups (site_id, name, description) VALUES 
(@cloud_site_id, '소개', '회사 소개 관련 페이지'),
(@cloud_site_id, '서비스', '서비스 관련 페이지');

-- 클라우드 사이트 메뉴 샘플 데이터
SET @cloud_intro_group_id = (SELECT group_id FROM page_groups WHERE name = '소개' AND site_id = @cloud_site_id);
SET @cloud_service_group_id = (SELECT group_id FROM page_groups WHERE name = '서비스' AND site_id = @cloud_site_id);

-- 클라우드 사이트 메뉴 샘플 데이터
INSERT INTO pages (site_id, group_id, title, slug, parent_id, depth, menu_order) VALUES 
(@cloud_site_id, @cloud_intro_group_id, '소개', 'intro', NULL, 0, 1),
(@cloud_site_id, @cloud_service_group_id, '서비스', 'service', NULL, 0, 2);

-- 클라우드 서비스의 하위 메뉴
SET @cloud_service_id = (SELECT page_id FROM pages WHERE slug = 'service' AND site_id = @cloud_site_id);

INSERT INTO pages (site_id, title, slug, parent_id, depth, menu_order) VALUES 
(@cloud_site_id, '상품소개', 'products', @cloud_service_id, 1, 1),
(@cloud_site_id, '가격', 'pricing', @cloud_service_id, 1, 2);

-- 스토어 사이트 메뉴 샘플 데이터
INSERT INTO pages (site_id, title, slug, parent_id, depth, content, menu_order) VALUES 
((SELECT site_id FROM sites WHERE code = 'store'), '홈', 'home', NULL, 0, '홈', 1),
((SELECT site_id FROM sites WHERE code = 'store'), '카테고리', 'categories', NULL, 0, '카테고리', 2);

-- 스토어 카테고리 하위 메뉴
SET @store_site_id = (SELECT site_id FROM sites WHERE code = 'store');
SET @store_categories_id = (SELECT page_id FROM pages WHERE slug = 'categories' AND site_id = @store_site_id);

INSERT INTO pages (site_id, title, slug,
 parent_id, depth, menu_order) VALUES 
(@store_site_id, '의류', 'clothing', @store_categories_id, 1, 1),
(@store_site_id, '전자기기', 'electronics', @store_categories_id, 1, 2);