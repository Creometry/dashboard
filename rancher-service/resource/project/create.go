package project

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// Exportable functions

func CreateProject(req ReqData) (data RespDataCreateProjectAndRepo, err error) {
	// create gitRepo
	repoName, err := createGitRepo(req.GitRepoName, req.GitRepoUrl, req.GitRepoBranch)
	if err != nil {
		return RespDataCreateProjectAndRepo{}, err
	}
	fmt.Printf("Created repo : %s", repoName)

	// create rancher project
	projectId, err := createRancherProject(req.UsrProjectName, req.Plan)
	if err != nil {
		return RespDataCreateProjectAndRepo{}, err
	}

	// add user to project
	_, err = AddUserToProject(req.UserId, projectId)
	if err != nil {
		return RespDataCreateProjectAndRepo{}, err
	}

	// make post request to resources-service/namespace and pass the project name and id to create a namespace in the specific project
	nsName, err := createNamespace(req.UsrProjectName, projectId)

	if err != nil {
		return RespDataCreateProjectAndRepo{}, err
	}
	// login as user to get token
	token, err := loginAsUser(req.UserId, "testtesttest")

	if err != nil {
		return RespDataCreateProjectAndRepo{}, err
	}

	resp := RespDataCreateProjectAndRepo{
		User_token: token,
		Namespace:  nsName,
		ProjectId:  projectId,
	}
	return resp, nil

}

func GetNamespaceByAnnotation(annotations []string) (string, string, error) {

	// http get request to get the namespace list with http client
	req, err := http.NewRequest("GET", os.Getenv("NAMESPACE_URL"), nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("RANCHER_TOKEN")))

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	// parse response body
	dt := RespDataNs{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	err = json.Unmarshal(body, &dt)
	if err != nil {
		return "", "", err
	}
	for _, annotation := range annotations {
		newAnnotation := fmt.Sprintf("%s:%s", os.Getenv("CLUSTER_ID"), strings.Split(annotation, ":")[0])
		for _, ns := range dt.Data {
			if ns.Metadata.Annotations["field.cattle.io/projectId"] == newAnnotation {
				return ns.Id, newAnnotation, nil
			}
		}
	}

	return "", "", nil

}

// TO DO : need to find an endpoint to authenticate user
// with his github code (if user not found in rancher, will be created)
func AuthFromCode(code string) (string, error) {
	return "", nil
}

func GetKubeConfig(token string) (string, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://tn.cloud.creometry.com/v3/clusters/%s?action=generateKubeconfig", os.Getenv("CLUSTER_ID")), nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	// parse response body
	dt := Kubeconfig{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(body, &dt)
	if err != nil {
		log.Fatal(err)
	}

	return dt.Config, nil
}

func AddUserToProject(userId string, projectId string) (RespDataRoleBinding, error) {

	req, err := http.NewRequest("POST", os.Getenv("ADD_USER_TO_PROJECT_URL"), bytes.NewBuffer([]byte(fmt.Sprintf(`{"userId":"%s","projectId":"%s","roleTemplateId":"project-member"}`, userId, projectId))))
	if err != nil {
		return RespDataRoleBinding{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("RANCHER_TOKEN")))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return RespDataRoleBinding{}, err
	}

	defer resp.Body.Close()
	// parse response body
	dt := RespDataRoleBinding{}
	body, err := ioutil.ReadAll(resp.Body)
	log.Print(string(body))
	if err != nil {
		return RespDataRoleBinding{}, err
	}
	err = json.Unmarshal(body, &dt)
	if err != nil {
		return RespDataRoleBinding{}, err
	}
	return dt, nil

}

func GetUserByUsername(username string) (string, []string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", os.Getenv("FIND_USER_URL"), username), nil)
	if err != nil {
		return "", []string{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("RANCHER_TOKEN")))

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return "", []string{}, err
	}

	defer resp.Body.Close()

	dt := FindUserData{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", []string{}, err
	}
	err = json.Unmarshal(body, &dt)
	if err != nil {
		return "", []string{}, err
	}

	if len(dt.Data) == 0 {
		return "", []string{}, errors.New("user not found")
	}
	return dt.Data[0].Id, dt.Data[0].PrincipalIds, nil

}

func ListTeamMembers(projectId string) ([]RespDataUserByUserId, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", os.Getenv("GET_PROJECT_MEMBERS_URL"), projectId), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("RANCHER_TOKEN")))

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	dt := RespDataTeamMembers{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &dt)
	if err != nil {
		return nil, err
	}
	var res []RespDataUserByUserId
	if len(dt.Data) > 0 {
		// loop through all the members, get their userId and get their names
		for _, user := range dt.Data {
			d, err := getUserById(strings.Split(user.UserId, "/")[0])
			if err != nil {
				continue
			} else {
				if d.Type != "error" {
					res = append(res, d)
				}
			}
		}
	}

	return res, nil

}

func FindUser(username string) (RespDataUser, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", os.Getenv("FIND_USER_URL"), username), nil)
	if err != nil {
		return RespDataUser{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("RANCHER_TOKEN")))

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return RespDataUser{}, err
	}
	defer resp.Body.Close()

	// parse response body
	dt := UserData{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return RespDataUser{}, err
	}
	err = json.Unmarshal(body, &dt)
	if err != nil {
		return RespDataUser{}, err
	}

	// if user exists, login and return token
	if len(dt.Data) > 0 {
		token, err := loginAsUser(username, "testtesttest")
		if err != nil {
			return RespDataUser{}, err
		}

		// get his projectName
		pr, err := getProjectsOfUser(dt.Data[0].Id, dt.Data[0].PrincipalIds)
		if err != nil {
			return RespDataUser{}, err
		}

		// get namespace of project
		if len(pr) > 0 {
			rs, prId, err := GetNamespaceByAnnotation(pr)
			if err != nil {
				return RespDataUser{}, err
			}

			log.Printf("rs: %s", rs)
			log.Printf("prId: %s", prId)
			return RespDataUser{
				Id:        dt.Data[0].Id,
				Token:     token,
				Namespace: rs,
				ProjectId: strings.Split(prId, ":")[1],
			}, nil
		} else {
			return RespDataUser{
				Id:        dt.Data[0].Id,
				Token:     token,
				Namespace: "",
				ProjectId: "",
			}, nil
		}

	}

	// if user does not exist, create user and return token
	id, _, err := createUser(username)
	if err != nil {
		return RespDataUser{}, err
	}

	token, err := loginAsUser(username, "testtesttest")
	if err != nil {
		return RespDataUser{}, err
	}
	return RespDataUser{
		Id:        id,
		Token:     token,
		Namespace: "",
	}, nil
}

// Local functions

func createRancherProject(usrProjectName string, plan string) (string, error) {
	resourceQuota := genResourceQuotaFromPlan(plan)
	if resourceQuota == "nil" {
		return "", fmt.Errorf("invalid plan")
	}
	req, err := http.NewRequest("POST", os.Getenv("CREATE_PROJECT_URL"), bytes.NewBuffer([]byte(fmt.Sprintf(`{"name":"%s","clusterId":"%s",%s}`, usrProjectName, os.Getenv("CLUSTER_ID"), resourceQuota))))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("RANCHER_TOKEN")))

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	// parse response body
	dt := RespData{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(body, &dt)
	if err != nil {
		log.Fatal(err)
	}

	return dt.ProjectId, nil
}

func genResourceQuotaFromPlan(plan string) string {
	switch plan {
	case "Starter":
		return `"namespaceDefaultResourceQuota": {
			"limit": {
			"configMaps": "10",
			"limitsCpu": "1000m",
			"limitsMemory": "2000Mi",
			"persistentVolumeClaims": "10",
			"pods": "50",
			"replicationControllers": "15",
			"requestsStorage": "50000Mi",
			"secrets": "20",
			"services": "50",
			"servicesLoadBalancers": "0",
			"servicesNodePorts": "0"
			}
			},
			"resourceQuota": {
			"limit": {
			"configMaps": "10",
			"limitsCpu": "1000m",
			"limitsMemory": "2000Mi",
			"persistentVolumeClaims": "10",
			"pods": "100",
			"replicationControllers": "30",
			"requestsStorage": "50000Mi",
			"secrets": "20",
			"services": "50",
			"servicesLoadBalancers": "0",
			"servicesNodePorts": "0"
			},
			"usedLimit": { }
			}
		`
	case "Pro":
		return `"namespaceDefaultResourceQuota": {
			"limit": {
			"configMaps": "20",
			"limitsCpu": "2000m",
			"limitsMemory": "4000Mi",
			"persistentVolumeClaims": "20",
			"pods": "100",
			"replicationControllers": "25",
			"requestsStorage": "50000Mi",
			"secrets": "20",
			"services": "50",
			"servicesLoadBalancers": "0",
			"servicesNodePorts": "0"
			}
			},
			"resourceQuota": {
			"limit": {
			"configMaps": "20",
			"limitsCpu": "2000m",
			"limitsMemory": "4000Mi",
			"persistentVolumeClaims": "20",
			"pods": "100",
			"replicationControllers": "25",
			"requestsStorage": "50000Mi",
			"secrets": "20",
			"services": "50",
			"servicesLoadBalancers": "0",
			"servicesNodePorts": "0"
			},
			"usedLimit": { }
			}
		`
	case "Elite":
		return `"namespaceDefaultResourceQuota": {
			"limit": {
			"configMaps": "20",
			"limitsCpu": "4000m",
			"limitsMemory": "8000Mi",
			"persistentVolumeClaims": "30",
			"pods": "200",
			"replicationControllers": "50",
			"requestsStorage": "200000Mi",
			"secrets": "20",
			"services": "100",
			"servicesLoadBalancers": "0",
			"servicesNodePorts": "0"
			}
			},
			"resourceQuota": {
			"limit": {
			"configMaps": "20",
			"limitsCpu": "4000m",
			"limitsMemory": "8000Mi",
			"persistentVolumeClaims": "30",
			"pods": "200",
			"replicationControllers": "50",
			"requestsStorage": "200000Mi",
			"secrets": "20",
			"services": "100",
			"servicesLoadBalancers": "0",
			"servicesNodePorts": "0"
			},
			"usedLimit": { }
			}
		`
	}
	return "nil"
}

func createUser(username string) (string, []string, error) {
	// should check if user exists before creating (TODO)
	userId, prIds, err := GetUserByUsername(username)
	if err != nil {
		return "", []string{}, err
	}

	if userId != "" {
		return userId, prIds, nil
	} else {
		req, err := http.NewRequest("POST", os.Getenv("CREATE_USER_URL"), bytes.NewBuffer([]byte(fmt.Sprintf(`{"username":"%s","mustChangePassword": false,"password": "testtesttest","principalIds": [ ]}`, username))))
		if err != nil {
			return "", []string{}, err
		}

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("RANCHER_TOKEN")))

		client := &http.Client{}

		resp, err := client.Do(req)

		if err != nil {
			return "", []string{}, err
		}

		defer resp.Body.Close()

		// parse response body
		dt := RespDataCreateUser{}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", []string{}, err
		}
		err = json.Unmarshal(body, &dt)
		if err != nil {
			return "", []string{}, err
		}
		log.Println("---------------")
		log.Println(dt)

		return dt.Id, dt.PrincipalIds, nil
	}

}

func loginAsUser(username string, password string) (string, error) {
	req, err := http.NewRequest("POST", os.Getenv("LOGIN_USER_URL"), bytes.NewBuffer([]byte(fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password))))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// parse response body
	dt := RespDataLogin{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(body, &dt)
	if err != nil {
		return "", err
	}

	return dt.Token, nil

}

func getUserById(userId string) (RespDataUserByUserId, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", os.Getenv("GET_USER_BY_ID"), userId), nil)
	if err != nil {
		return RespDataUserByUserId{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("RANCHER_TOKEN")))

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return RespDataUserByUserId{}, err
	}

	defer resp.Body.Close()

	// parse response body
	dt := RespDataUserByUserId{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return RespDataUserByUserId{}, err
	}
	err = json.Unmarshal(body, &dt)
	if err != nil {
		return RespDataUserByUserId{}, err
	}

	return dt, nil

}

func createGitRepo(name string, url string, branch string) (string, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("https://tn.cloud.creometry.com/k8s/clusters/%s/v1/catalog.cattle.io.clusterrepos", os.Getenv("CLUSTER_ID")), bytes.NewBuffer([]byte(fmt.Sprintf(`{
		"type": "catalog.cattle.io.clusterrepo",
		"metadata": {
		  "name": "%s"
		},
		"spec": {
		  "url": "",
		  "clientSecret": null,
		  "gitRepo": "%s",
		  "gitBranch": "%s"
		}
	  }`, name, url, branch))))

	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("RANCHER_TOKEN")))

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// parse response body
	dt := RespDataCreateGitRepo{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(body, &dt)
	if err != nil {
		return "", err
	}

	return dt.Id, nil
}

func getProjectsOfUser(userId string, principalIds []string) ([]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", os.Getenv("GET_USER_PROJECTS"), userId), nil)
	if err != nil {
		return []string{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("RANCHER_TOKEN")))

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	// parse response body
	dt := RespDataProjectsByUser{}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []string{}, err
	}
	err = json.Unmarshal(body, &dt)
	if err != nil {
		return []string{}, err
	}

	log.Println(dt)

	if len(dt.Data) > 0 {
		// return all the ids
		res := []string{}
		for _, v := range dt.Data {
			res = append(res, v.Id)
		}
		return res, nil

	}

	return []string{}, nil
}

func createNamespace(projectName string, projectId string) (string, error) {

	req, err := http.NewRequest("POST", os.Getenv("CREATE_NAMESPACE_URL"), bytes.NewBuffer([]byte(fmt.Sprintf(`{"projectName":"%s","projectId":"%s"}`, projectName, projectId))))

	if err != nil {
		return "", err
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// parse response body

	dt := CreateNsRespData{}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &dt)

	if err != nil {
		return "", err
	}

	if dt.Error != "" {
		return "", errors.New(dt.Error)
	}

	return dt.NsName, nil

}