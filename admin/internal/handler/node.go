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

var defaultNodeRouter = new(NodeRouter)

// @Summary 删除节点
// @Description 根据UUID删除节点。节点必须处于非活动状态
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

// @Summary 搜索节点
// @Description 根据指定条件搜索节点
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
