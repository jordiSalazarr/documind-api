import { z } from 'zod';

export const ItemTypeSchema = z.object({
  id: z.number().int().positive(),
  name: z.string().min(1).max(100),
  projectId: z.number().int().positive(),
  createdAt: z.date(),
  updatedAt: z.date(),
});

export const CreateItemTypeSchema = ItemTypeSchema.pick({
  name: true,
  projectId: true,
});

export const UpdateItemTypeSchema = ItemTypeSchema.pick({
  name: true,
}).partial();

export type ItemType = z.infer<typeof ItemTypeSchema>;
export type CreateItemTypeDto = z.infer<typeof CreateItemTypeSchema>;
export type UpdateItemTypeDto = z.infer<typeof UpdateItemTypeSchema>;
