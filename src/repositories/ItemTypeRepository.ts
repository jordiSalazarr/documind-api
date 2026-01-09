import { db } from '../db/db';
import { itemTypes } from '../db/schema';
import { eq, and } from 'drizzle-orm';
import type { ItemType, CreateItemTypeDto, UpdateItemTypeDto } from '../models/ItemType';

export class ItemTypeRepository {
  async findAll(projectId: number): Promise<ItemType[]> {
    return db.select().from(itemTypes).where(eq(itemTypes.projectId, projectId));
  }

  async findById(id: number, projectId: number): Promise<ItemType | undefined> {
    const results = await db
      .select()
      .from(itemTypes)
      .where(and(eq(itemTypes.id, id), eq(itemTypes.projectId, projectId)))
      .limit(1);
    return results[0];
  }

  async create(dto: CreateItemTypeDto): Promise<ItemType> {
    const results = await db.insert(itemTypes).values(dto).returning();
    return results[0];
  }

  async update(id: number, projectId: number, dto: UpdateItemTypeDto): Promise<ItemType | undefined> {
    const results = await db
      .update(itemTypes)
      .set({
        ...dto,
        updatedAt: new Date(),
      })
      .where(and(eq(itemTypes.id, id), eq(itemTypes.projectId, projectId)))
      .returning();
    return results[0];
  }

  async delete(id: number, projectId: number): Promise<boolean> {
    const results = await db
      .delete(itemTypes)
      .where(and(eq(itemTypes.id, id), eq(itemTypes.projectId, projectId)))
      .returning();
    return results.length > 0;
  }

  async existsByName(name: string, projectId: number, excludeId?: number): Promise<boolean> {
    const conditions = excludeId
      ? and(eq(itemTypes.name, name), eq(itemTypes.projectId, projectId), eq(itemTypes.id, excludeId))
      : and(eq(itemTypes.name, name), eq(itemTypes.projectId, projectId));

    const results = await db.select({ id: itemTypes.id }).from(itemTypes).where(conditions).limit(1);
    return results.length > 0;
  }
}
