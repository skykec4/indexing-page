import { RowDataPacket } from "mysql2";

export interface Site extends RowDataPacket {
  site_id: number;
  code: string;
  name: string;
  domain: string;
  created_at: Date;
  updated_at: Date;
}

export interface PageGroup extends RowDataPacket {
  group_id: number;
  site_id: number;
  name: string;
  description: string;
  created_at: Date;
  updated_at: Date;
  menu: PageTree[]; // 각 그룹 아래에 메뉴가 있음
}

export interface Page {
  id: number;
  site_id: number;
  group_id: number | null;
  title: string;
  slug: string;
  parent_id: number | null;
  depth: number;
  menu_order: number;
  content: string;
  is_published: boolean;
  created_at: Date;
  updated_at: Date;
}

export interface PageTree extends Page {
  children: PageTree[];
}

export interface SiteResponse extends Site {
  pageGroups: PageGroup[];
}

export interface CreatePageInput {
  title: string;
  slug: string;
  parent_id?: number;
  group_id?: number;
  content: string;
}

export interface UpdatePageInput {
  title: string;
  slug: string;
  content: string;
  group_id?: number;
}

export interface ISite {
  code: string;
  name: string;
  domain: string;
}

export interface IPageGroup {
  name: string;
  description?: string;
}
