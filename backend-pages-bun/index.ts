import { serve } from "bun";
import { pool } from "./src/config/database";
import {
  Page,
  PageTree,
  CreatePageInput,
  UpdatePageInput,
  SiteResponse,
  Site,
  ISite,
  PageGroup,
  IPageGroup,
} from "./src/models/page";
import { RowDataPacket } from "mysql2/promise";
import {
  getCodeList,
  getSiteList,
  getPageGroups,
  createPageGroup,
  updatePageGroup,
  deletePageGroup,
} from "./src/handler/handler";

const server = serve({
  port: 3001,
  routes: {
    "/api/status": new Response("OK"),
    "/api/sites": {
      GET: async () => {
        const sites = await getSiteList();
        return Response.json(sites, { status: 200 });
      },
      POST: async (req: Request) => {
        const body = (await req.json()) as ISite;
        const [result] = await pool.query(
          `
            INSERT INTO sites (code, name, domain)
            VALUES (?, ?, ?)
          `,
          [body.code, body.name, body.domain]
        );
        return Response.json({ created: true, site_id: result.insertId });
      },
    },
    "/api/menu/:code": {
      GET: async (req) => {
        try {
          const [rows] = await pool.query<Site[]>(
            `
              SELECT site_id, code, name, domain, created_at, updated_at
              FROM sites
              WHERE code = ?
            `,
            [req.params.code]
          );

          if (rows.length === 0) {
            return Response.json(
              { error: "사이트를 찾을 수 없습니다" },
              { status: 404 }
            );
          }
          console.log("row::", rows);

          // 그룹 조회
          const [groups] = await pool.query<PageGroup[]>(
            `
              SELECT pg.group_id, pg.site_id, pg.name, pg.description, pg.created_at, pg.updated_at
              FROM page_groups pg
              WHERE pg.site_id = ?
              ORDER BY pg.name
            `,
            [rows[0].site_id]
          );

          // 각 그룹의 메뉴 조회
          const groupMenus = await Promise.all(
            groups.map(async (group) => {
              const [rows2] = await pool.query<RowDataPacket[]>(
                `
                  SELECT p.id, p.site_id, p.group_id, p.title, p.slug, p.parent_id, p.depth,
                         p.menu_order, p.content, p.is_published, p.created_at, p.updated_at
                  FROM pages p
                  WHERE p.site_id = ? AND p.group_id = ?
                  ORDER BY p.depth, p.menu_order
                `,
                [rows[0].site_id, group.group_id]
              );

              return {
                ...group,
                menu: buildMenuTree(
                  rows2.map((row) => ({
                    id: row.id,
                    site_id: row.site_id,
                    group_id: row.group_id,
                    title: row.title,
                    slug: row.slug,
                    parent_id: row.parent_id,
                    depth: row.depth,
                    menu_order: row.menu_order,
                    content: row.content,
                    is_published: row.is_published,
                    created_at: row.created_at,
                    updated_at: row.updated_at,
                  }))
                ),
              };
            })
          );

          const siteResponse: SiteResponse = {
            ...rows[0],
            pageGroups: groupMenus,
          };
          return new Response(JSON.stringify(siteResponse || null), {
            headers: { "Content-Type": "application/json" },
          });
        } catch (error) {
          console.log("error : ", error);
          return Response.json(
            { error: "요청 처리 중 오류가 발생했습니다." },
            { status: 500 }
          );
        }
      },
      // POST 요청 개선 예시
      POST: async (req) => {
        try {
          const body = await req.json();

          // 요청 검증
          if (!body.title || !body.slug) {
            return new Response(
              JSON.stringify({ error: "Missing required fields" }),
              {
                status: 400,
              }
            );
          }

          // 데이터베이스 삽입
          const [result] = await pool.query(
            `
      INSERT INTO pages (site_id, title, slug, parent_id, depth, menu_order, content, is_published)
      VALUES (?, ?, ?, ?, ?, ?, ?, ?)
      `,
            [
              body.site_id,
              body.title,
              body.slug,
              body.parent_id || null,
              body.parent_id ? 1 : 0,
              0,
              body.content || "",
              true,
            ]
          );

          // 생성된 페이지 조회
          const [rows] = await pool.query(`SELECT * FROM pages WHERE id = ?`, [
            result.insertId,
          ]);

          return Response.json(rows[0], { status: 201 });
        } catch (error) {
          console.error("Error creating page:", error);
          return new Response(
            JSON.stringify({ error: "Failed to create page" }),
            {
              status: 500,
            }
          );
        }
      },
    },
    "/api/sites/:siteId/groups": {
      GET: async (req) => {
        const siteId = parseInt(req.params.siteId);
        const groups = await getPageGroups(siteId);
        return Response.json(groups, { status: 200 });
      },
      POST: async (req) => {
        const siteId = parseInt(req.params.siteId);
        const body = (await req.json()) as IPageGroup;
        const groupId = await createPageGroup(siteId, body);
        return Response.json({ created: true, group_id: groupId });
      },
    },
    "/api/sites/:siteId/groups/:group_id": {
      PUT: async (req) => {
        const group_id = parseInt(req.params.group_id);
        const body = (await req.json()) as Partial<IPageGroup>;
        await updatePageGroup(group_id, body);
        return Response.json({ updated: true });
      },
      DELETE: async (req) => {
        const group_id = parseInt(req.params.group_id);
        await deletePageGroup(group_id);
        return Response.json({ deleted: true });
      },
    },
  },
  async fetch(req) {
    const url = new URL(req.url);
    const siteCode = url.searchParams.get("siteCode");
    const pageId = url.searchParams.get("pageId");

    console.log(siteCode);

    // 404 처리
    return new Response("페이지를 찾을 수 없습니다", { status: 404 });
    // return new Response("");
  },
});

function buildMenuTree(pages: Page[]): PageTree[] {
  const pageMap = new Map<number, PageTree>();
  const roots: PageTree[] = [];

  // 모든 페이지를 트리 노드로 변환
  pages.forEach((page) => {
    pageMap.set(page.id, { ...page, children: [] });
  });

  // 트리 구조 구성
  pages.forEach((page) => {
    const treeNode = pageMap.get(page.id)!;
    if (!page.parent_id) {
      roots.push(treeNode);
    } else {
      const parent = pageMap.get(page.parent_id);
      if (parent) {
        parent.children.push(treeNode);
      }
    }
  });

  return roots;
}

console.log(`서버가 http://localhost:${server.port} 에서 실행 중입니다`);
