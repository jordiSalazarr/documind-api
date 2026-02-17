package knowledge

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"documind.jordi.org/internal/shared/infrastructure/middleware"

	// Project
	projectcreate "documind.jordi.org/internal/knowledge/project/features/commands/create"
	projectdelete "documind.jordi.org/internal/knowledge/project/features/commands/delete"
	projectupdate "documind.jordi.org/internal/knowledge/project/features/commands/update"
	projectget "documind.jordi.org/internal/knowledge/project/features/queries/get"
	projectlist "documind.jordi.org/internal/knowledge/project/features/queries/list"

	// Area
	areacreate "documind.jordi.org/internal/knowledge/area/features/commands/create"
	areadelete "documind.jordi.org/internal/knowledge/area/features/commands/delete"
	areaupdate "documind.jordi.org/internal/knowledge/area/features/commands/update"
	areaget "documind.jordi.org/internal/knowledge/area/features/queries/get"
	arealist "documind.jordi.org/internal/knowledge/area/features/queries/list"

	// Item
	itemcreate "documind.jordi.org/internal/knowledge/item/features/commands/create"
	itemcreateversion "documind.jordi.org/internal/knowledge/item/features/commands/createversion"
	itemupload "documind.jordi.org/internal/knowledge/item/features/commands/upload"
	itemdelete "documind.jordi.org/internal/knowledge/item/features/commands/delete"
	itempublish "documind.jordi.org/internal/knowledge/item/features/commands/publish"
	itemupdate "documind.jordi.org/internal/knowledge/item/features/commands/update"
	itemget "documind.jordi.org/internal/knowledge/item/features/queries/get"
	itemgetversion "documind.jordi.org/internal/knowledge/item/features/queries/getversion"
	itemlist "documind.jordi.org/internal/knowledge/item/features/queries/list"
	itemlistversions "documind.jordi.org/internal/knowledge/item/features/queries/listversions"

	// ItemType
	itemtypecreate "documind.jordi.org/internal/knowledge/itemtype/features/commands/create"
	itemtypedelete "documind.jordi.org/internal/knowledge/itemtype/features/commands/delete"
	itemtypeupdate "documind.jordi.org/internal/knowledge/itemtype/features/commands/update"
	itemtypeget "documind.jordi.org/internal/knowledge/itemtype/features/queries/get"
	itemtypelist "documind.jordi.org/internal/knowledge/itemtype/features/queries/list"

	// Relation
	"documind.jordi.org/internal/knowledge/relation/features/commands/createrelation"
	"documind.jordi.org/internal/knowledge/relation/features/commands/createtype"
	"documind.jordi.org/internal/knowledge/relation/features/commands/deleterelation"
	"documind.jordi.org/internal/knowledge/relation/features/commands/deletetype"
	"documind.jordi.org/internal/knowledge/relation/features/commands/updatetype"
	"documind.jordi.org/internal/knowledge/relation/features/queries/getrelation"
	"documind.jordi.org/internal/knowledge/relation/features/queries/gettype"
	"documind.jordi.org/internal/knowledge/relation/features/queries/listrelations"
	"documind.jordi.org/internal/knowledge/relation/features/queries/listtypes"
	"documind.jordi.org/internal/knowledge/relation/persistence"
)

// Deps holds external dependencies needed by the knowledge context.
type Deps struct {
	DB            *sql.DB
	Chunker       itemcreate.DocumentChunker
	FileParser    itemupload.FileParser    // optional: for file upload
	FileValidator itemupload.FileValidator // optional: for file upload
}

// RegisterRoutes registers all knowledge context routes under /knowledge group.
func RegisterRoutes(rg *gin.RouterGroup, deps Deps) {
	kg := rg.Group("/knowledge")
	editor := middleware.RequireRole("editor")

	// Project routes
	projects := kg.Group("/projects")
	{
		projects.POST("", editor, projectcreate.Endpoint(projectcreate.NewHandler(deps.DB)))
		projects.GET("", projectlist.Endpoint(projectlist.NewHandler(deps.DB)))
		projects.GET("/:projectId", projectget.Endpoint(projectget.NewHandler(deps.DB)))
		projects.PUT("/:projectId", editor, projectupdate.Endpoint(projectupdate.NewHandler(deps.DB)))
		projects.DELETE("/:projectId", editor, projectdelete.Endpoint(projectdelete.NewHandler(deps.DB)))

		// Area routes (nested under project)
		areas := projects.Group("/:projectId/areas")
		{
			areas.POST("", editor, areacreate.Endpoint(areacreate.NewHandler(deps.DB)))
			areas.GET("", arealist.Endpoint(arealist.NewHandler(deps.DB)))
			areas.GET("/:id", areaget.Endpoint(areaget.NewHandler(deps.DB)))
			areas.PUT("/:id", editor, areaupdate.Endpoint(areaupdate.NewHandler(deps.DB)))
			areas.DELETE("/:id", editor, areadelete.Endpoint(areadelete.NewHandler(deps.DB)))
		}
	}

	// Item routes
	items := kg.Group("/items")
	{
		items.POST("", editor, itemcreate.Endpoint(itemcreate.NewHandler(deps.DB, deps.Chunker)))
		items.GET("", itemlist.Endpoint(itemlist.NewHandler(deps.DB)))
		items.GET("/:id", itemget.Endpoint(itemget.NewHandler(deps.DB)))
		items.PUT("/:id", editor, itemupdate.Endpoint(itemupdate.NewHandler(deps.DB)))
		items.DELETE("/:id", editor, itemdelete.Endpoint(itemdelete.NewHandler(deps.DB)))
		items.POST("/:id/publish", editor, itempublish.Endpoint(itempublish.NewHandler(deps.DB)))
		items.POST("/:id/versions", editor, itemcreateversion.Endpoint(itemcreateversion.NewHandler(deps.DB, deps.Chunker)))
		items.GET("/:id/versions", itemlistversions.Endpoint(itemlistversions.NewHandler(deps.DB)))
		items.GET("/:id/versions/:versionId", itemgetversion.Endpoint(itemgetversion.NewHandler(deps.DB)))

		// File upload endpoint
		if deps.FileParser != nil && deps.FileValidator != nil {
			uploadHandler := itemupload.NewHandler(deps.DB, deps.FileParser, deps.FileValidator, deps.Chunker)
			items.POST("/upload", editor, itemupload.Endpoint(uploadHandler))
		}
	}

	// ItemType routes
	itemTypes := kg.Group("/item-types")
	{
		itemTypes.POST("", editor, itemtypecreate.Endpoint(itemtypecreate.NewHandler(deps.DB)))
		itemTypes.GET("", itemtypelist.Endpoint(itemtypelist.NewHandler(deps.DB)))
		itemTypes.GET("/:id", itemtypeget.Endpoint(itemtypeget.NewHandler(deps.DB)))
		itemTypes.PUT("/:id", editor, itemtypeupdate.Endpoint(itemtypeupdate.NewHandler(deps.DB)))
		itemTypes.DELETE("/:id", editor, itemtypedelete.Endpoint(itemtypedelete.NewHandler(deps.DB)))
	}

	// Relation routes
	typeRepo := &persistence.RelationTypeRepo{DB: deps.DB}
	relRepo := &persistence.RelationRepo{DB: deps.DB}

	relationTypes := kg.Group("/relation-types")
	{
		relationTypes.POST("", editor, createtype.Endpoint(createtype.NewHandler(typeRepo)))
		relationTypes.GET("/:id", gettype.Endpoint(gettype.NewHandler(typeRepo)))
		relationTypes.GET("", listtypes.Endpoint(listtypes.NewHandler(typeRepo)))
		relationTypes.PUT("/:id", editor, updatetype.Endpoint(updatetype.NewHandler(typeRepo)))
		relationTypes.DELETE("/:id", editor, deletetype.Endpoint(deletetype.NewHandler(typeRepo)))
	}

	relations := kg.Group("/relations")
	{
		relations.POST("", editor, createrelation.Endpoint(createrelation.NewHandler(typeRepo, relRepo)))
		relations.GET("/:id", getrelation.Endpoint(getrelation.NewHandler(relRepo)))
		relations.GET("", listrelations.Endpoint(listrelations.NewHandler(relRepo)))
		relations.DELETE("/:id", editor, deleterelation.Endpoint(deleterelation.NewHandler(relRepo)))
	}
}
