package controllers

import (
	"fmt"
	"os"
	"strings"

	"github.com/Creometry/dashboard/go-provisioner/internal/project"
	"github.com/Creometry/dashboard/go-provisioner/internal/team"
	"github.com/gofiber/fiber/v2"
)

func ProvisionProject(c *fiber.Ctx) error {
	// parse the request body
	reqData := new(project.ReqData)
	if err := c.BodyParser(reqData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	// check if the request body is valid
	if err := reqData.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	data, err := project.ProvisionProject(*reqData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"projectId": data.ProjectId,
	})
}

func ProvisionProjectNewUser(c *fiber.Ctx) error {
	// parse the request body
	reqData := new(project.ReqDataNewUser)
	if err := c.BodyParser(reqData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	// check if the request body is valid
	if err := reqData.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	data, err := project.ProvisionProjectNewUser(*reqData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"projectId": data.ProjectId,
		"token":     data.Token,
		"password":  data.Password,
	})
}

func GenerateKubeConfig(c *fiber.Ctx) error {
	// get the token from the body
	reqData := new(project.ReqDataKubeconfig)
	if err := c.BodyParser(reqData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	// check if the request body is valid
	if reqData.Token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "token is required",
		})
	}
	data, err := project.GetKubeConfig(reqData.Token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"config": data,
	})
}

// func FindUserAndLoginOrCreate(c *fiber.Ctx) error {
// 	// get username from path
// 	username := c.Params("username")
// 	if username == "" {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": "username is required",
// 		})
// 	}

// 	data, err := project.FindUser(username)
// 	if err != nil {
// 		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
// 			"error": err.Error(),
// 		})
// 	}
// 	return c.JSON(fiber.Map{
// 		"user_token": data.Token,
// 		"user_id":    data.Id,
// 		"namespace":  data.Namespace,
// 		"projectId":  data.ProjectId,
// 	})
// }

func ListTeamMembers(c *fiber.Ctx) error {
	projectId := c.Params("projectId")
	if projectId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "projectId is required",
		})
	}
	// if projectId contains ':' then list team members without change
	var prId string
	if strings.Contains(projectId, ":") {
		prId = projectId
	} else {
		prId = fmt.Sprintf("%s:%s", os.Getenv("CLUSTER_ID"), projectId)
	}
	data, err := team.ListTeamMembers(prId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if data == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no members found or invalid projectId",
		})
	}
	return c.JSON(fiber.Map{
		"members": data,
		"prId":    prId,
	})
}

func AddTeamMember(c *fiber.Ctx) error {
	projectId := c.Params("projectId")
	userId := c.Params("userId")

	if projectId == "" || userId == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "projectId and userId are required",
		})
	}

	var prId string
	if strings.Contains(projectId, ":") {
		prId = projectId
	} else {
		prId = fmt.Sprintf("%s:%s", os.Getenv("CLUSTER_ID"), projectId)
	}

	data, err := project.AddUserToProject(userId, prId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": data,
	})
}

func Login(c *fiber.Ctx) error {
	// get the token from the body
	reqData := new(project.ReqDataLogin)
	if err := c.BodyParser(reqData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	// check if the request body is valid
	if err := reqData.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	token, err := project.Login(reqData.Username, reqData.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.JSON(fiber.Map{
		"token": token,
	})
}
