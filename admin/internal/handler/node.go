package handler

import (
	"fmt"
	"pulse/admin/internal/model/request"
	"pulse/admin/internal/model/resp"
	"pulse/admin/internal/service"
	"pulse/common/models"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// NodeRouter 结构体用于定义相关的路由处理器
type NodeRouter struct{}

// defaultNodeRouter 是 NodeRouter 的一个全局单例，在注册路由时使用
var defaultNodeRouter = new(NodeRouter)

// @Summary Delete a node
// @Description Deletes a node by its UUID. The node must be inactive.
// @Tags node
// @Accept  json
// @Produce  json
// @Param   body  body   request.ByUUID  true  "Node UUID"
// @Success 200 {object} resp.Response "Delete success"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /node/del [post]
func (n *NodeRouter) Delete(c *gin.Context) {
	var req request.ByUUID
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_node] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[delete_node] request parameter error", c)
	}

	node := &models.Node{UUID: req.UUID}
	err := node.FindByUUID()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_node] find node by uuid: %s error: %s", req.UUID, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[deleta_node db find error]", c)
		return
	}

	// can't delete a node that is already alive
	if node.Status == models.NodeConnSuccess {
		resp.FailWithMessage(resp.ERROR, "[delete_node] can't delete a node that is already alive", c)
		return
	}

	// 使用 `WithPrefix()` 可以删除 /pulse/job/profile/<node_uuid>/ 前缀下的所有键值对
	_, _ = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdJobProfile, req.UUID), clientv3.WithPrefix())

	err = node.Delete()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_node] into db error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[delete_node] db delete error", c)
		return
	}

	resp.OkWithMessage("delete success", c)
}

// @Summary Search for nodes
// @Description Searches for nodes based on specified criteria.
// @Tags node
// @Accept  json
// @Produce  json
// @Param   body  body   request.ReqNodeSearch  true  "Search criteria"
// @Success 200 {object} resp.PageResult "Search results"
// @Failure 400 {object} resp.Response "Bad request"
// @Failure 500 {object} resp.Response "Internal server error"
// @Router /node/search [post]
func (n *NodeRouter) Search(c *gin.Context) {
	var req request.ReqNodeSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_node] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[search_node] request parameter error", c)
		return
	}
	req.Check()

	nodes, total, err := service.DefaultNodeWatcher.Search(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_node] search node error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[search_node] search node error", c)
		return
	}

	var resultNodes []resp.RspNodeSearch
	for _, node := range nodes {
		resultNode := resp.RspNodeSearch{
			Node: node,
		}
		resultNode.JobCount, _ = service.DefaultNodeWatcher.GetJobCount(node.UUID)
		resultNodes = append(resultNodes, resultNode)
	}

	resp.OkWithDetailed(resp.PageResult{
		List:     resultNodes,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "search success", c)
}
