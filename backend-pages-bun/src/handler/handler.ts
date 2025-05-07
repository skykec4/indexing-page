import { RowDataPacket } from "mysql2";
import { pool } from "../config/database";
import {
  Site,
  PageGroup,
  Page,
  SiteResponse,
  ISite,
  IPageGroup,
} from "../models/page";

export const getSiteList = async () => {
  try {
    const sql = `
      SELECT site_id, code, name, domain, created_at, updated_at
      FROM sites
    `;
    const [rows] = await pool.query<Site[]>({
      sql,
    });

    return rows;
  } catch (error) {
    throw error;
  }
};

export const getCodeList = async () => {
  try {
    const [rowsCode] = await pool.query<RowDataPacket[]>(
      `
        SELECT code
        FROM sites
      `
    );

    const codeList = rowsCode.map((row) => row.code);

    return codeList;
  } catch (error) {
    throw error;
  }
};

export const getPageGroups = async (siteId: number) => {
  try {
    const sql = `
      SELECT group_id, site_id, name, description, created_at, updated_at
      FROM page_groups
      WHERE site_id = ?
      ORDER BY name
    `;
    const [rows] = await pool.query<PageGroup[]>({
      sql,
      values: [siteId],
    });

    return rows;
  } catch (error) {
    throw error;
  }
};

export const createPageGroup = async (siteId: number, group: IPageGroup) => {
  try {
    const sql = `
      INSERT INTO page_groups (site_id, name, description)
      VALUES (?, ?, ?)
    `;
    const [result] = await pool.query({
      sql,
      values: [siteId, group.name, group.description],
    });

    return result.insertId;
  } catch (error) {
    throw error;
  }
};

export const updatePageGroup = async (
  groupId: number,
  group: Partial<IPageGroup>
) => {
  try {
    const sql = `
      UPDATE page_groups
      SET name = ?, description = ?
      WHERE group_id = ?
    `;
    await pool.query({
      sql,
      values: [group.name, group.description, groupId],
    });
  } catch (error) {
    throw error;
  }
};

export const deletePageGroup = async (groupId: number) => {
  try {
    const sql = `
      DELETE FROM page_groups WHERE group_id = ?
    `;
    await pool.query({
      sql,
      values: [groupId],
    });
  } catch (error) {
    throw error;
  }
};
