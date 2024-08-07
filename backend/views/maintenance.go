package views

import (
	"net/http"

	"github.com/AdamPekny/IIS/backend/models"
	"github.com/AdamPekny/IIS/backend/serializers"
	"github.com/AdamPekny/IIS/backend/utils"
	"github.com/gin-gonic/gin"
)

// MALFUNCTION REPORT

func CreateMalfuncReport(ctx *gin.Context) {
	var malfunc_report_serializer serializers.MalfuncRepCreateSerialzier

	if err := ctx.BindJSON(&malfunc_report_serializer); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	malfunc_report_serializer.CreatedByRef = &logged_user.ID

	malfunc_report_model := malfunc_report_serializer.ToModel()

	result := utils.DB.Create(malfunc_report_model)

	if result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	if result := utils.DB.First(&malfunc_report_model.CreatedBy, malfunc_report_model.CreatedByRef); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	if result := utils.DB.Where("registration = ?", malfunc_report_model.VehicleRef).First(&malfunc_report_model.Vehicle); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	var malfunc_report_pub_serializer serializers.MalfuncRepPublicSerialzier

	if err := malfunc_report_pub_serializer.FromModel(malfunc_report_model); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, malfunc_report_pub_serializer)
}

func ListStatusMalfuncReports(ctx *gin.Context) {
	var malfunc_reports []models.MalfunctionReport
	var malfunc_report_serializers []serializers.MalfuncRepPublicSerialzier

	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
	}

	db_query := utils.DB

	if logged_user.Role == models.DriverRole {
		db_query = db_query.Where("created_by_ref = ?", logged_user.ID)
	}

	vehicle := ctx.Query("vehicle")
	if vehicle != "" {
		db_query = db_query.Where("vehicle_ref = ?", vehicle)
	}

	if result := db_query.Preload("MaintenReqs").Preload("CreatedBy").Preload("Vehicle").Find(&malfunc_reports); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	status := ctx.Query("status")

	for _, report := range malfunc_reports {
		if (status == "ack" && len(report.MaintenReqs) > 0) || (status == "unack" && len(report.MaintenReqs) == 0) || (status != "ack" && status != "unack") {
			malfunc_report_serializers = append(malfunc_report_serializers, serializers.MalfuncRepPublicSerialzier{})
			if err := malfunc_report_serializers[len(malfunc_report_serializers)-1].FromModel(&report); err != nil {
				ctx.IndentedJSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}
		}
	}

	ctx.IndentedJSON(http.StatusOK, malfunc_report_serializers)
}

func GetMalfuncReport(ctx *gin.Context) {
	id, err := utils.GetIDFromURL(ctx)

	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var malfunc_report models.MalfunctionReport
	var malfunc_report_serializer serializers.MalfuncRepPublicSerialzier

	if result := utils.DB.Preload("CreatedBy").Preload("Vehicle").First(&malfunc_report, id); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	if err := malfunc_report_serializer.FromModel(&malfunc_report); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, malfunc_report_serializer)
}

func UpdateMalfuncReport(ctx *gin.Context) {
	id, err := utils.GetIDFromURL(ctx)

	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var original_report_model models.MalfunctionReport

	if result := utils.DB.First(&original_report_model, id); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	if logged_user.ID != *original_report_model.CreatedByRef && logged_user.Role != models.AdminRole {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": "permission denied",
		})
		return
	}

	var new_report_serializer serializers.MalfuncRepCreateSerialzier

	if err := ctx.BindJSON(&new_report_serializer); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	original_report_model.Title = new_report_serializer.Title
	original_report_model.Description = new_report_serializer.Description
	original_report_model.VehicleRef = new_report_serializer.VehicleRef

	result := utils.DB.Save(original_report_model)

	if result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	if result := utils.DB.First(&original_report_model.CreatedBy, original_report_model.CreatedByRef); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	if result := utils.DB.Where("registration = ?", original_report_model.VehicleRef).First(&original_report_model.Vehicle); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	var new_report_pub_serializer serializers.MalfuncRepPublicSerialzier

	if err := new_report_pub_serializer.FromModel(&original_report_model); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, new_report_pub_serializer)
}

func DeleteMalfuncReport(ctx *gin.Context) {
	id, err := utils.GetIDFromURL(ctx)

	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var report_model models.MalfunctionReport

	if result := utils.DB.First(&report_model, id); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	if logged_user.ID != *report_model.CreatedByRef && logged_user.Role != models.AdminRole {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": "permission denied",
		})
		return
	}

	if result := utils.DB.Delete(report_model); result.Error != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{
		"message": "malfunction report deleted successfully",
	})
}

// MAINTENANCE REQUEST

func CreateMaintenRequest(ctx *gin.Context) {
	var mainten_req_serializer serializers.MaintenReqCreateSerializer

	if err := ctx.BindJSON(&mainten_req_serializer); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !mainten_req_serializer.Valid() {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": mainten_req_serializer.ValidatorErrs,
		})
		return
	}

	mainten_req_model, err := mainten_req_serializer.ToModel(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": err.Error(),
		})
		return
	}

	if result := utils.DB.Create(mainten_req_model); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": result.Error.Error(),
		})
		return
	}

	// Fill Malfunction report
	if result := utils.DB.First(&mainten_req_model.MalfuncRep, mainten_req_model.MalfuncRepRef); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": result.Error.Error(),
		})
		return
	}

	// Fill Created By
	if result := utils.DB.First(&mainten_req_model.CreatedBy, mainten_req_model.CreatedByRef); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": result.Error.Error(),
		})
		return
	}

	if mainten_req_model.ResolvedByRef != nil {
		// Fill Resolved By
		if result := utils.DB.First(&mainten_req_model.ResolvedBy, mainten_req_model.ResolvedByRef); result.Error != nil {
			ctx.IndentedJSON(http.StatusBadRequest, gin.H{
				"errors": result.Error.Error(),
			})
			return
		}
	}

	var mainten_req_pub_serializer serializers.MaintenReqPublicSerializer

	if err := mainten_req_pub_serializer.FromModel(mainten_req_model); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, mainten_req_pub_serializer)
}

func ListUnassignedMaintenRequests(ctx *gin.Context) {
	var mainten_req_models []models.MaintenanceRequest
	var mainten_req_model_serializers []serializers.MaintenReqPublicSerializer

	if result := utils.DB.Where("resolved_by_ref IS NULL").Preload("MalfuncRep").Preload("CreatedBy").Preload("ResolvedBy").Find(&mainten_req_models); result.Error != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	for _, request := range mainten_req_models {
		mainten_req_model_serializers = append(mainten_req_model_serializers, serializers.MaintenReqPublicSerializer{})

		if err := mainten_req_model_serializers[len(mainten_req_model_serializers)-1].FromModel(&request); err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	ctx.IndentedJSON(http.StatusOK, mainten_req_model_serializers)
}

func ListStatusMaintenRequests(ctx *gin.Context) {
	var mainten_req_models []models.MaintenanceRequest
	var mainten_req_model_serializers []serializers.MaintenReqPublicSerializer

	db_query := utils.DB

	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	if logged_user.Role == models.TechnicianRole {
		db_query = db_query.Where("resolved_by_ref = ?", logged_user.ID)
	}
	
	status := ctx.Query("status")
	if status == string(models.PendingStatus) || status == string(models.InProgressStatus) || status == string(models.DoneStatus) {
		db_query = db_query.Where("status = ?", status)
	}

	vehicle := ctx.Query("vehicle")
	if vehicle != "" {
		db_query = db_query.Joins("JOIN malfunction_reports ON malfunction_reports.id = maintenance_requests.malfunc_rep_ref").Where("malfunction_reports.vehicle_ref = ?", vehicle)
	}

	if result := db_query.Preload("MalfuncRep").Preload("CreatedBy").Preload("ResolvedBy").Find(&mainten_req_models); result.Error != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	for _, request := range mainten_req_models {
		mainten_req_model_serializers = append(mainten_req_model_serializers, serializers.MaintenReqPublicSerializer{})

		if err := mainten_req_model_serializers[len(mainten_req_model_serializers)-1].FromModel(&request); err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	ctx.IndentedJSON(http.StatusOK, mainten_req_model_serializers)
}

func ListCreatorStatusMaintenRequests(ctx *gin.Context) {
	var mainten_req_models []models.MaintenanceRequest
	var mainten_req_model_serializers []serializers.MaintenReqPublicSerializer


	id, err := utils.GetIDFromURL(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	status := ctx.Param("status")

	if status == "all" {
		if result := utils.DB.Where("created_by_ref = ?", id).Preload("MalfuncRep").Preload("CreatedBy").Preload("ResolvedBy").Find(&mainten_req_models); result.Error != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
				"error": result.Error.Error(),
			})
			return
		}
	} else {
		if result := utils.DB.Where("status = ? AND created_by_ref = ?", status, id).Preload("MalfuncRep").Preload("CreatedBy").Preload("ResolvedBy").Find(&mainten_req_models); result.Error != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
				"error": result.Error.Error(),
			})
			return
		}
	}

	for _, request := range mainten_req_models {
		mainten_req_model_serializers = append(mainten_req_model_serializers, serializers.MaintenReqPublicSerializer{})

		if err := mainten_req_model_serializers[len(mainten_req_model_serializers)-1].FromModel(&request); err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	ctx.IndentedJSON(http.StatusOK, mainten_req_model_serializers)
}

func ListResolverStatusMaintenRequests(ctx *gin.Context) {
	var mainten_req_models []models.MaintenanceRequest
	var mainten_req_model_serializers []serializers.MaintenReqPublicSerializer


	id, err := utils.GetIDFromURL(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	status := ctx.Param("status")

	if status == "all" {
		if result := utils.DB.Where("resolved_by_ref = ?", id).Preload("MalfuncRep").Preload("CreatedBy").Preload("ResolvedBy").Find(&mainten_req_models); result.Error != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
				"error": result.Error.Error(),
			})
			return
		}
	} else {
		if result := utils.DB.Where("status = ? AND resolved_by_ref = ?", status, id).Preload("MalfuncRep").Preload("CreatedBy").Preload("ResolvedBy").Find(&mainten_req_models); result.Error != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
				"error": result.Error.Error(),
			})
			return
		}
	}

	for _, request := range mainten_req_models {
		mainten_req_model_serializers = append(mainten_req_model_serializers, serializers.MaintenReqPublicSerializer{})

		if err := mainten_req_model_serializers[len(mainten_req_model_serializers)-1].FromModel(&request); err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	ctx.IndentedJSON(http.StatusOK, mainten_req_model_serializers)
}

func GetMaintenRequest(ctx *gin.Context) {
	var mainten_req_model models.MaintenanceRequest
	var mainten_req_model_serializer serializers.MaintenReqPublicSerializer


	id, err := utils.GetIDFromURL(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if result := utils.DB.Preload("MalfuncRep").Preload("CreatedBy").Preload("ResolvedBy").First(&mainten_req_model, id); result.Error != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	if err := mainten_req_model_serializer.FromModel(&mainten_req_model); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, mainten_req_model_serializer)
}

func UpdateMaintenRequest(ctx *gin.Context) {
	var mainten_req_serializer serializers.MaintenReqUpdateSerializer

	if err := ctx.BindJSON(&mainten_req_serializer); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !mainten_req_serializer.Valid() {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": mainten_req_serializer.ValidatorErrs,
		})
		return
	}

	mainten_req_model, err := mainten_req_serializer.ToModel(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": err.Error(),
		})
		return
	}

	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if logged_user.ID != *mainten_req_model.CreatedByRef && logged_user.Role != models.AdminRole {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": "permission denied",
		})
		return
	}

	if result := utils.DB.Save(mainten_req_model); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": result.Error.Error(),
		})
		return
	}

	// Fill Malfunction report
	if result := utils.DB.First(&mainten_req_model.MalfuncRep, mainten_req_model.MalfuncRepRef); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": result.Error.Error(),
		})
		return
	}

	// Fill Created By
	if result := utils.DB.First(&mainten_req_model.CreatedBy, mainten_req_model.CreatedByRef); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": result.Error.Error(),
		})
		return
	}

	if mainten_req_model.ResolvedByRef != nil {
		// Fill Resolved By
		if result := utils.DB.First(&mainten_req_model.ResolvedBy, mainten_req_model.ResolvedByRef); result.Error != nil {
			ctx.IndentedJSON(http.StatusBadRequest, gin.H{
				"errors": result.Error.Error(),
			})
			return
		}
	}

	var mainten_req_pub_serializer serializers.MaintenReqPublicSerializer

	if err := mainten_req_pub_serializer.FromModel(mainten_req_model); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, mainten_req_pub_serializer)
}

func AssignTechMaintenRequest(ctx *gin.Context) {
	var mainten_req_serializer serializers.MaintenReqAssignTechSerializer

	if err := ctx.BindJSON(&mainten_req_serializer); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !mainten_req_serializer.Valid() {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": mainten_req_serializer.ValidatorErrs,
		})
		return
	}

	mainten_req_model, err := mainten_req_serializer.ToModel(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": err.Error(),
		})
		return
	}

	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if logged_user.Role == models.TechnicianRole && logged_user.ID != *mainten_req_model.ResolvedByRef {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": "permission denied",
		})
		return
	} else if logged_user.Role == models.SuperuserRole && logged_user.ID != *mainten_req_model.CreatedByRef {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": "permission denied",
		})
		return
	}

	if result := utils.DB.Model(mainten_req_model).Update("resolved_by_ref", mainten_req_model.ResolvedByRef); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	// Fill with new data
	if result := utils.DB.Preload("MalfuncRep").Preload("CreatedBy").Preload("ResolvedBy").First(&mainten_req_model, mainten_req_model.ID); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	var mainten_req_pub_serializer serializers.MaintenReqPublicSerializer

	if err := mainten_req_pub_serializer.FromModel(mainten_req_model); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, mainten_req_pub_serializer)
}

func DeleteMaintenRequest(ctx *gin.Context) {
	id, err := utils.GetIDFromURL(ctx)

	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var request_model models.MaintenanceRequest

	if result := utils.DB.First(&request_model, id); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	if logged_user.ID != *request_model.CreatedByRef && logged_user.Role != models.AdminRole {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": "permission denied",
		})
		return
	}

	if result := utils.DB.Delete(request_model); result.Error != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{
		"message": "maintenance request deleted successfully",
	})
}

func CreateMaintenReport(ctx *gin.Context) {
	report_create_serializer := &serializers.MaintenRepCreateSerializer{}
	if err := ctx.BindJSON(report_create_serializer); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !report_create_serializer.Valid() {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": report_create_serializer.ValidatorErrs,
		})
		return
	}

	request_model := &models.MaintenanceRequest{}
	if result := utils.DB.First(request_model, report_create_serializer.MaintenReqRef); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	if logged_user.ID != *request_model.ResolvedByRef && logged_user.Role != models.AdminRole {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": "permission denied",
		})
		return
	}

	report_model, err := report_create_serializer.ToModel()
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	if result := utils.DB.Create(report_model); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	report_public_serializer := &serializers.MaintenRepPublicSerializer{}
	if err := report_public_serializer.FromModel(report_model); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, report_public_serializer)
}

func ListMaintenReports(ctx *gin.Context) {
	var mainten_rep_models []models.MaintenanceReport
	var mainten_rep_pub_serializers []serializers.MaintenRepPublicSerializer
	
	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	db_query := utils.DB

	if logged_user.Role == models.TechnicianRole {
		db_query = db_query.Joins("JOIN maintenance_requests ON maintenance_reports.mainten_req_ref = maintenance_requests.id").Where("maintenance_requests.resolved_by_ref = ?", logged_user.ID)
	}

	if result := db_query.Order("created_at DESC").Find(&mainten_rep_models); result.Error != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	for _, report := range mainten_rep_models {
		mainten_rep_pub_serializers = append(mainten_rep_pub_serializers, serializers.MaintenRepPublicSerializer{})
		if err := mainten_rep_pub_serializers[len(mainten_rep_pub_serializers)-1].FromModel(&report); err != nil {
			ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	ctx.IndentedJSON(http.StatusOK, mainten_rep_pub_serializers)
}

func GetMaintenReport(ctx *gin.Context) {
	var mainten_rep_model models.MaintenanceReport
	var mainten_rep_pub_serializer serializers.MaintenRepPublicSerializer

	id, err := utils.GetIDFromURL(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if result := utils.DB.First(&mainten_rep_model, id); result.Error != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	if err := mainten_rep_pub_serializer.FromModel(&mainten_rep_model); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, mainten_rep_pub_serializer)
}

func UpdateMaintenReport(ctx *gin.Context) {
	report_update_serializer := &serializers.MaintenRepUpdateSerializer{}
	if err := ctx.BindJSON(report_update_serializer); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if !report_update_serializer.Valid() {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"errors": report_update_serializer.ValidatorErrs,
		})
		return
	}

	request_model := &models.MaintenanceRequest{}
	if result := utils.DB.First(request_model, report_update_serializer.MaintenReqRef); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	if logged_user.ID != *request_model.ResolvedByRef && logged_user.Role != models.AdminRole {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": "permission denied",
		})
		return
	}

	report_model, err := report_update_serializer.ToModel(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	if result := utils.DB.Save(report_model); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	report_public_serializer := &serializers.MaintenRepPublicSerializer{}
	if err := report_public_serializer.FromModel(report_model); err != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, report_public_serializer)
}

func DeleteMaintenReport(ctx *gin.Context) {
	id, err := utils.GetIDFromURL(ctx)

	if err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var report_model models.MaintenanceReport

	if result := utils.DB.First(&report_model, id); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	request_model := &models.MaintenanceRequest{}
	if result := utils.DB.First(request_model, report_model.MaintenReqRef); result.Error != nil {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	logged_user, err := models.GetUserFromCtx(ctx)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	if logged_user.ID != *request_model.ResolvedByRef && logged_user.Role != models.AdminRole {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"error": "permission denied",
		})
		return
	}

	if result := utils.DB.Delete(report_model); result.Error != nil {
		ctx.IndentedJSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{
		"message": "maintenance report deleted successfully",
	})
}
