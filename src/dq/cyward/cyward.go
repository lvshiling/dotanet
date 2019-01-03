package cyward

import (
	"dq/log"
	"dq/vec2d"
	"time"
)

type DetourPathNode struct {
	parent      *DetourPathNode
	collions    *Body
	my          *Body
	serachIndex int
	path1len    float64
	path1       []vec2d.Vec2
	path2       []vec2d.Vec2
}

type Body struct {
	Core           *WardCore
	Position       vec2d.Vec2   //当前位置
	R              vec2d.Vec2   //矩形半径
	SpeedSize      float64      //移动速度大小
	TargetPosition []vec2d.Vec2 //移动目标位置
	DetourPath     []vec2d.Vec2 //绕路 路径

	CollisoinStopTime float64 //碰撞停止移动剩余时间
	CurSpeedSize      float64 //当前速度大小

	TargetIndex  int        //计算后的目标位置索引
	NextPosition vec2d.Vec2 //计算后的下一帧位置
	Direction    vec2d.Vec2 //速度方向

	Tag int //标记
}

//避障核心
type WardCore struct {
	Bodys []*Body
}

func (this *Body) SetTag(tag int) {
	this.Tag = tag
}
func (this *Body) Update(dt float64) {

	this.CollisoinStopTime -= dt
	if this.CalcNextPosition(dt) {

		//log.Info("nextposition x:%f y:%f", this.NextPosition.X, this.NextPosition.Y)

		//检查碰撞
		collisionOne := this.CheckPositionCollisoin(dt)
		if collisionOne != nil {
			//log.Info("collisionOne:%d", collisionOne.Tag)
			if collisionOne.CurSpeedSize > 0 {
				this.CollisoinStopTime = 0.5
				this.CurSpeedSize = 0
				this.NextPosition = this.Position
				this.TargetIndex = 0
			} else {

				this.Core.CalcDetourPath(this, collisionOne, this.TargetPosition[0], &this.DetourPath)
				if len(this.DetourPath) <= 0 {
					this.TargetPosition = this.TargetPosition[1:]
				}

			}
		} else {
			this.Position = this.NextPosition
			this.CurSpeedSize = this.SpeedSize

			for i := 0; i < this.TargetIndex; i++ {
				if len(this.DetourPath) > 0 {
					this.DetourPath = this.DetourPath[1:]
				} else {
					this.TargetPosition = this.TargetPosition[1:]
				}
			}

			//log.Info("DetourPathlen:%d", len(this.DetourPath))
		}

	}

}

func (this *Body) SetTarget(pos vec2d.Vec2) {

	log.Info("SetTarget %f  %f", pos.X, pos.Y)

	this.TargetPosition = this.TargetPosition[0:0]
	this.DetourPath = this.DetourPath[0:0]
	this.TargetPosition = append(this.TargetPosition, pos)
	this.CollisoinStopTime = 0

	t1 := time.Now().UnixNano()

	dpNode := &DetourPathNode{}
	dpNode.parent = nil
	dpNode.collions = nil
	dpNode.my = this
	dpNode.serachIndex = 0
	dpNode.path1 = make([]vec2d.Vec2, 0)
	dpNode.path1 = append(dpNode.path1, this.Position)
	dpNode.path1 = append(dpNode.path1, pos)

	bodys := make([]*Body, 0)
	this.Core.GetStaticBodys(&bodys)
	if this.Core.CheckDetourPathNodeT(dpNode, &bodys, &this.DetourPath) {
		log.Info("SetTarget %d", this.Tag)
		for i := 0; i < len(this.DetourPath); i++ {
			log.Info("x: %f  y:%f", this.DetourPath[i].X, this.DetourPath[i].Y)
		}
	} else {
		//cocos2d::log("SetTarget 222");
	}

	t2 := time.Now().UnixNano()
	log.Info("time:%d", (t2-t1)/1e6)
}

func (this *Body) CheckPositionCollisoin(dt float64) *Body {
	return this.Core.GetNextPositionCollision(this)
}

func (this *Body) IsCollisionPoint(p vec2d.Vec2) bool {
	if this.Position.X-this.R.X <= p.X && p.X <= this.Position.X+this.R.X &&
		this.Position.Y-this.R.Y <= p.Y && p.Y <= this.Position.Y+this.R.Y {
		return true
	}
	return false
}

//获取目标位置
func (this *Body) GetTargetPos(index int, pos *vec2d.Vec2) bool {
	if index < len(this.DetourPath) {
		*pos = this.DetourPath[index]
		return true
	} else {
		if index < len(this.DetourPath)+len(this.TargetPosition) {

			*pos = this.TargetPosition[index-len(this.DetourPath)]
			return true
		} else {
			return false
		}
	}

}

func (this *Body) CalcNextPosition(dt float64) bool {
	if (len(this.TargetPosition) <= 0 && len(this.DetourPath) <= 0) || this.CollisoinStopTime > 0 {
		this.CurSpeedSize = 0
		this.NextPosition = this.Position
		return false
	}
	//log.Info("CalcNextPosition tag:%d", this.Tag)

	//var startpos vec2d.Vec2
	startpos := this.Position

	var targetpos vec2d.Vec2
	this.GetTargetPos(0, &targetpos)
	//目标方向
	speeddir := vec2d.Sub(targetpos, this.Position)
	//cocos2d::Vec2 speeddir = targetpos - Position;
	//剩余到目标点的距离
	targetdis := speeddir.Length()
	//移动距离
	movedis := this.SpeedSize * dt
	this.TargetIndex = 0
	//log.Info("targetdis:%f  movedis:%f ", targetdis, movedis)
	//while () {
	for {
		if targetdis >= movedis {
			break
		}

		this.TargetIndex++
		if this.TargetIndex >= len(this.TargetPosition)+len(this.DetourPath) {
			this.NextPosition = targetpos
			this.Direction = speeddir.GetNormalized()
			return true
		} else {
			startpos = targetpos
			movedis = movedis - targetdis

			this.GetTargetPos(this.TargetIndex, &targetpos)
			var lastpos vec2d.Vec2
			this.GetTargetPos(this.TargetIndex-1, &lastpos)
			speeddir = vec2d.Sub(targetpos, lastpos)
			targetdis = speeddir.Length()
		}
		//log.Info("11targetdis:%f  movedis:%f ", targetdis, movedis)

	}
	this.Direction = speeddir.GetNormalized()
	this.NextPosition = vec2d.Add(startpos, vec2d.Mul(this.Direction, movedis))

	return true
}

//线段是否与矩形相交
func (this *WardCore) IsSegmentCollionSquare(p1 vec2d.Vec2, p2 vec2d.Vec2, pCenter vec2d.Vec2, r vec2d.Vec2) bool {
	//变成正方形
	circlep1 := vec2d.Add(pCenter, vec2d.Vec2{-r.X, r.Y})
	circlep2 := vec2d.Add(pCenter, vec2d.Vec2{-r.X, -r.Y})
	circlep3 := vec2d.Add(pCenter, vec2d.Vec2{r.X, -r.Y})
	circlep4 := vec2d.Add(pCenter, vec2d.Vec2{r.X, r.Y})

	//判断线段是否与线段相交

	if vec2d.IsSegmentIntersect(p1, p2, circlep1, circlep2) || vec2d.IsSegmentIntersect(p1, p2, circlep2, circlep3) || vec2d.IsSegmentIntersect(p1, p2, circlep3, circlep4) || vec2d.IsSegmentIntersect(p1, p2, circlep4, circlep1) {
		return true
	} else {
		return false
	}

	return false

}
func (this *WardCore) GetIntersectPoint(A vec2d.Vec2, B vec2d.Vec2, C vec2d.Vec2, D vec2d.Vec2, Re *vec2d.Vec2) bool {
	var S, T float64

	//if (cocos2d::Vec2::isLineIntersect(A, B, C, D, &S, &T))
	if vec2d.IsLineIntersect(A, B, C, D, &S, &T) && (S >= 0.0 && S <= 1.0 && T >= 0.0 && T <= 1.0) {
		// Vec2 of intersection
		//cocos2d::Vec2 P;
		Re.X = A.X + S*(B.X-A.X)
		Re.Y = A.Y + S*(B.Y-A.Y)
		return true
	}

	return false
}
func (this *WardCore) GetSegmentInsterset(p1 vec2d.Vec2, p2 vec2d.Vec2, pCenter vec2d.Vec2, r vec2d.Vec2, Re *vec2d.Vec2) bool {
	//变成正方形
	circlep1 := vec2d.Add(pCenter, vec2d.Vec2{-r.X, r.Y})
	circlep2 := vec2d.Add(pCenter, vec2d.Vec2{-r.X, -r.Y})
	circlep3 := vec2d.Add(pCenter, vec2d.Vec2{r.X, -r.Y})
	circlep4 := vec2d.Add(pCenter, vec2d.Vec2{r.X, r.Y})

	//判断线段是否与线段相交

	if this.GetIntersectPoint(p1, p2, circlep1, circlep2, Re) {
		return true
	} else if this.GetIntersectPoint(p1, p2, circlep2, circlep3, Re) {
		return true
	} else if this.GetIntersectPoint(p1, p2, circlep3, circlep4, Re) {
		return true
	} else if this.GetIntersectPoint(p1, p2, circlep4, circlep1, Re) {
		return true
	} else {
		return false
	}
}

func (this *WardCore) GetStaticBodys(bodys *[]*Body) {
	for i := 0; i < len(this.Bodys); i++ {
		if this.Bodys[i].CurSpeedSize <= 0 {
			(*bodys) = append((*bodys), this.Bodys[i])
		}
	}
}

func (this *WardCore) GetBodys() *[]*Body {
	return &this.Bodys
}
func (this *WardCore) GetPointIndexFromSquare(centerPoint vec2d.Vec2, r vec2d.Vec2, targetPos vec2d.Vec2, posIndex *[]int) {
	//正方形的4个顶点
	var points [4]vec2d.Vec2
	points[0] = vec2d.Vec2{centerPoint.X - r.X, centerPoint.Y + r.Y}
	points[1] = vec2d.Vec2{centerPoint.X + r.X, centerPoint.Y + r.Y}
	points[2] = vec2d.Vec2{centerPoint.X + r.X, centerPoint.Y - r.Y}
	points[3] = vec2d.Vec2{centerPoint.X - r.X, centerPoint.Y - r.Y}

	if targetPos.X <= points[0].X && targetPos.Y >= points[0].Y { //目标点在矩形的左上
		(*posIndex) = append((*posIndex), 3)
		(*posIndex) = append((*posIndex), 0)
		(*posIndex) = append((*posIndex), 1)
	} else if targetPos.X >= points[0].X && targetPos.X <= points[1].X && targetPos.Y >= points[0].Y { //正上
		(*posIndex) = append((*posIndex), 0)
		(*posIndex) = append((*posIndex), 1)
	} else if targetPos.X >= points[1].X && targetPos.Y >= points[0].Y { //右上
		(*posIndex) = append((*posIndex), 0)
		(*posIndex) = append((*posIndex), 1)
		(*posIndex) = append((*posIndex), 2)
	} else if targetPos.X >= points[1].X && targetPos.Y < points[1].Y && targetPos.Y >= points[2].Y { //正右
		(*posIndex) = append((*posIndex), 1)
		(*posIndex) = append((*posIndex), 2)
	} else if targetPos.X >= points[1].X && targetPos.Y <= points[2].Y { //右下
		(*posIndex) = append((*posIndex), 1)
		(*posIndex) = append((*posIndex), 2)
		(*posIndex) = append((*posIndex), 3)
	} else if targetPos.X >= points[3].X && targetPos.X <= points[2].X && targetPos.Y <= points[2].Y { //正下
		(*posIndex) = append((*posIndex), 2)
		(*posIndex) = append((*posIndex), 3)
	} else if targetPos.X <= points[3].X && targetPos.Y <= points[2].Y { //左下
		(*posIndex) = append((*posIndex), 2)
		(*posIndex) = append((*posIndex), 3)
		(*posIndex) = append((*posIndex), 0)
	} else if targetPos.X <= points[3].X && targetPos.Y <= points[0].Y && targetPos.Y >= points[3].Y { //正左
		(*posIndex) = append((*posIndex), 3)
		(*posIndex) = append((*posIndex), 0)
	}
}

func (this *WardCore) GetLen(path []vec2d.Vec2) float64 {
	if len(path) < 2 {
		return 0
	}
	re := 0.0
	for i := 0; i < len(path)-1; i++ {
		v1 := vec2d.Sub(path[i+1], path[i])
		re += v1.Length()
	}
	return re
}

//计算绕行路径
func (this *WardCore) CalcDetourPathFromSquare(p1 vec2d.Vec2, centerPoint vec2d.Vec2, r vec2d.Vec2, targetPos vec2d.Vec2, path1 *[]vec2d.Vec2, path2 *[]vec2d.Vec2) bool {
	//目标点在正方形内部
	if centerPoint.X-r.X <= targetPos.X && centerPoint.X+r.X >= targetPos.X &&
		centerPoint.Y-r.Y <= targetPos.Y && centerPoint.Y+r.Y >= targetPos.Y {
		return false
	}

	if centerPoint.X-r.X <= p1.X && centerPoint.X+r.X >= p1.X &&
		centerPoint.Y-r.Y <= p1.Y && centerPoint.Y+r.Y >= p1.Y {
		return false
	}
	//r = r + 1;
	//正方形的4个顶点
	var points [4]vec2d.Vec2
	points[0] = vec2d.Vec2{centerPoint.X - r.X, centerPoint.Y + r.Y}
	points[1] = vec2d.Vec2{centerPoint.X + r.X, centerPoint.Y + r.Y}
	points[2] = vec2d.Vec2{centerPoint.X + r.X, centerPoint.Y - r.Y}
	points[3] = vec2d.Vec2{centerPoint.X - r.X, centerPoint.Y - r.Y}
	//计算目标点能直接通过的顶点
	var points2TargetIndex []int
	var points2P1Index []int

	this.GetPointIndexFromSquare(centerPoint, r, targetPos, &points2TargetIndex)
	this.GetPointIndexFromSquare(centerPoint, r, p1, &points2P1Index)
	//删掉中间点(只保留两端顶点)
	if len(points2P1Index) >= 3 {
		points2P1Index = append(points2P1Index[:1], points2P1Index[2:]...)
	}
	rightp := points2P1Index[0]
	leftp := points2P1Index[1]

	//向外偏移一个像素
	var offset [4]vec2d.Vec2
	offset[0] = vec2d.Vec2{-1, 1}
	offset[1] = vec2d.Vec2{1, 1}
	offset[2] = vec2d.Vec2{1, -1}
	offset[3] = vec2d.Vec2{-1, -1}
	has := false
	for {
		if has == true {
			break
		}
		(*path1) = append((*path1), vec2d.Add(points[rightp], offset[rightp]))

		for i := 0; i < len(points2TargetIndex); i++ {
			if points2TargetIndex[i] == rightp {
				has = true
				break
			}
		}
		rightp -= 1
		if rightp < 0 {
			rightp = 3
		}
	}
	has = false
	for {
		if has == true {
			break
		}
		(*path2) = append((*path2), vec2d.Add(points[leftp], offset[leftp]))
		for i := 0; i < len(points2TargetIndex); i++ {
			if points2TargetIndex[i] == leftp {
				has = true
				break
			}
		}
		leftp += 1
		if leftp > 3 {
			leftp = 0
		}
	}
	return true
}

func (this *WardCore) ChangeErrorPath(my *Body, detourBody *Body, staticbodys *[]*Body, path1 *[]vec2d.Vec2, path2 *[]vec2d.Vec2) {
	for j := 0; j < len(*path1); j++ {
		for i := 0; i < len(*staticbodys); i++ {
			if (*staticbodys)[i] == my || (*staticbodys)[i] == detourBody {
				continue
			}
			R := vec2d.Add(my.R, (*staticbodys)[i].R)
			if (*staticbodys)[i].Position.X-R.X <= (*path1)[j].X && (*path1)[j].X <= (*staticbodys)[i].Position.X+R.X &&
				(*staticbodys)[i].Position.Y-R.Y <= (*path1)[j].Y && (*path1)[j].Y <= (*staticbodys)[i].Position.Y+R.Y {
				//更改点
				dir := vec2d.Sub((*path1)[j], detourBody.Position)
				seg := vec2d.Add(vec2d.Mul(dir.GetNormalized(), 10000), (*path1)[j])

				var intersectPoint vec2d.Vec2
				if (this.GetSegmentInsterset((*path1)[j], seg, (*staticbodys)[i].Position, vec2d.Add(R, vec2d.Vec2{1, 1}), &intersectPoint)) {
					(*path1)[j] = intersectPoint
					j--
					break
				}

			}
		}
	}
	for j := 0; j < len(*path2); j++ {
		for i := 0; i < len(*staticbodys); i++ {
			if (*staticbodys)[i] == my || (*staticbodys)[i] == detourBody {
				continue
			}
			R := vec2d.Add(my.R, (*staticbodys)[i].R)
			if (*staticbodys)[i].Position.X-R.X <= (*path2)[j].X && (*path2)[j].X <= (*staticbodys)[i].Position.X+R.X &&
				(*staticbodys)[i].Position.Y-R.Y <= (*path2)[j].Y && (*path2)[j].Y <= (*staticbodys)[i].Position.Y+R.Y {
				//更改点
				dir := vec2d.Sub((*path2)[j], detourBody.Position)
				seg := vec2d.Add(vec2d.Mul(dir.GetNormalized(), 10000), (*path2)[j])

				var intersectPoint vec2d.Vec2
				if (this.GetSegmentInsterset((*path2)[j], seg, (*staticbodys)[i].Position, vec2d.Add(R, vec2d.Vec2{1, 1}), &intersectPoint)) {
					(*path2)[j] = intersectPoint
					j--
					break
				}

			}
		}
	}
}

func (this *WardCore) CheckDetourPathNodeT(dpnode *DetourPathNode, staticbodys *[]*Body, path *[]vec2d.Vec2) bool {
	var getPath [2][]vec2d.Vec2
	this.CheckDetourPathNode1(dpnode, staticbodys, &getPath[0])
	this.OptimizePath(dpnode.my, staticbodys, &getPath[0])

	this.CheckDetourPathNode2(dpnode, staticbodys, &getPath[1])
	this.OptimizePath(dpnode.my, staticbodys, &getPath[1])

	if len(getPath[0]) <= 0 && len(getPath[1]) <= 0 {
		(*path) = make([]vec2d.Vec2, 0)
		//(*path) = path[0,0]
		return false
	}
	if len(getPath[0]) <= 0 {
		(*path) = make([]vec2d.Vec2, len(getPath[1]))
		copy((*path), getPath[1])
		return true
	}
	if len(getPath[1]) <= 0 {
		(*path) = make([]vec2d.Vec2, len(getPath[0]))
		copy((*path), getPath[0])
		return true
	}

	len1 := this.GetLen(getPath[0])
	len2 := this.GetLen(getPath[1])
	if len1 > len2 {
		(*path) = make([]vec2d.Vec2, len(getPath[1]))
		copy((*path), getPath[1])
		return true
	} else {
		(*path) = make([]vec2d.Vec2, len(getPath[0]))
		copy((*path), getPath[0])
		return true
	}
}
func (this *WardCore) OptimizePath(me *Body, staticbodys *[]*Body, path *[]vec2d.Vec2) {
	if len(*path) <= 2 {
		return
	}

	for start := 0; start < len(*path)-1; start++ {
		for end := len(*path) - 1; end > start; end-- {
			isCollion := false
			p1 := (*path)[start]
			p2 := (*path)[end]

			for i := 0; i < len(*staticbodys); i++ {
				//if (staticbodys[i] == dpnode->collions || staticbodys[i] == dpnode->my) {
				if (*staticbodys)[i] == me {
					continue
				}
				R := vec2d.Add((*staticbodys)[i].R, me.R)
				//
				if this.IsSegmentCollionSquare(p1, p2, (*staticbodys)[i].Position, R) {
					isCollion = true
					break
				}
			}
			if !isCollion {
				//删除点
				(*path) = append((*path)[:start+1], (*path)[end:]...)
				break
			}
		}
	}
}
func (this *WardCore) CheckDetourPathNode2(dpnode *DetourPathNode, staticbodys *[]*Body, path *[]vec2d.Vec2) bool {
	for k := 0; k < 2; k++ {
		dpnodepath1 := make([]vec2d.Vec2, 0)
		if k == 1 {
			dpnodepath1 = make([]vec2d.Vec2, len(dpnode.path1))
			copy(dpnodepath1, dpnode.path1)
		} else {
			dpnodepath1 = make([]vec2d.Vec2, len(dpnode.path2))
			copy(dpnodepath1, dpnode.path2)
		}
		if len(dpnodepath1) <= 0 {
			continue
		}
		//cocos2d::log("--------start--------------%d",k);

		canPassAblePath1 := true //路径1是否可以通行
		isbreakpath := false

		for pathindex := dpnode.serachIndex; pathindex < len(dpnodepath1)-1; pathindex++ {
			p1 := dpnodepath1[pathindex]
			p2 := dpnodepath1[pathindex+1]

			minDisSquared := 10000000000.0
			var minDisCollion *Body

			for i := 0; i < len(*staticbodys); i++ {
				//if (staticbodys[i] == dpnode->collions || staticbodys[i] == dpnode->my) {
				if (*staticbodys)[i] == dpnode.my {
					continue
				}
				R := vec2d.Add((*staticbodys)[i].R, dpnode.my.R)
				//
				if this.IsSegmentCollionSquare(p1, p2, (*staticbodys)[i].Position, R) {
					//继续绕路
					if (*staticbodys)[i] == dpnode.collions {
						log.Info("staticbodys[i] == dpnode->collions---%d", (*staticbodys)[i].Tag)
					}
					t1 := vec2d.Sub((*staticbodys)[i].Position, p1)
					disSquared := t1.LengthSquared()
					if minDisCollion == nil {
						minDisCollion = (*staticbodys)[i]
						minDisSquared = disSquared
					} else {
						if minDisSquared > disSquared {
							minDisCollion = (*staticbodys)[i]
							minDisSquared = disSquared
						}
					}

				} else {

				}
			}
			if minDisCollion != nil {
				//如果与之前的所有父节点有碰撞 则不能通行
				parent := dpnode.parent
				isbreak := false
				for {
					if parent == nil {
						break
					}
					if minDisCollion == parent.collions {
						isbreak = true
						break
					}
					parent = parent.parent
				}
				if isbreak {
					canPassAblePath1 = false
					isbreakpath = true
					break
				}
				R := vec2d.Add(minDisCollion.R, dpnode.my.R)

				path1 := make([]vec2d.Vec2, 0)
				path2 := make([]vec2d.Vec2, 0)

				detourPointIndex := pathindex + 1
				for {
					if detourPointIndex >= len(dpnodepath1) {
						canPassAblePath1 = false
						isbreakpath = true
						break
					}
					//log.Info("----DoCheckGameData--tag:%d", minDisCollion.Tag)
					//cocos2d::log("--------tag:%d", minDisCollion->Tag);

					if this.CalcDetourPathFromSquare(p1, minDisCollion.Position, R, dpnodepath1[detourPointIndex], &path1, &path2) {

						this.ChangeErrorPath(dpnode.my, minDisCollion, staticbodys, &path1, &path2)

						var dpNode1 DetourPathNode
						dpNode1.parent = dpnode
						dpNode1.collions = minDisCollion
						dpNode1.my = dpnode.my
						dpNode1.serachIndex = pathindex
						first := append([]vec2d.Vec2{}, dpnodepath1[:pathindex+1]...)
						first2 := append([]vec2d.Vec2{}, dpnodepath1[:pathindex+1]...)
						rear := append([]vec2d.Vec2{}, dpnodepath1[detourPointIndex:]...)
						rear2 := append([]vec2d.Vec2{}, dpnodepath1[detourPointIndex:]...)

						dpNode1.path1 = make([]vec2d.Vec2, 0)
						dpNode1.path1 = append(first, path1[:]...)
						dpNode1.path1 = append(dpNode1.path1, rear...)

						dpNode1.path2 = make([]vec2d.Vec2, 0)
						dpNode1.path2 = append(first2, path2[:]...)
						dpNode1.path2 = append(dpNode1.path2, rear2...)

						canPassAblePath1 = this.CheckDetourPathNode2(&dpNode1, staticbodys, path)
						if canPassAblePath1 == true {
							//log.Info("--------canPassAblePath1--------------")
							return true
						}
						isbreakpath = true
						break
					} else {
						//此目标点 不能绕行
						//canPassAblePath1 = false;

					}
					detourPointIndex++
				}

				break
			}
			if isbreakpath == true {
				break
			}

		}
		if canPassAblePath1 == true {
			(*path) = make([]vec2d.Vec2, len(dpnodepath1))
			copy((*path), dpnodepath1)
			return true
		} else {
			//return false;
		}

	}

	return false
}

func (this *WardCore) CheckDetourPathNode1(dpnode *DetourPathNode, staticbodys *[]*Body, path *[]vec2d.Vec2) bool {
	for k := 0; k < 2; k++ {
		dpnodepath1 := make([]vec2d.Vec2, 0)
		if k == 0 {
			dpnodepath1 = make([]vec2d.Vec2, len(dpnode.path1))
			copy(dpnodepath1, dpnode.path1)
		} else {
			dpnodepath1 = make([]vec2d.Vec2, len(dpnode.path2))
			copy(dpnodepath1, dpnode.path2)
		}
		if len(dpnodepath1) <= 0 {
			continue
		}
		//cocos2d::log("--------start--------------%d",k);

		canPassAblePath1 := true //路径1是否可以通行
		isbreakpath := false

		for pathindex := dpnode.serachIndex; pathindex < len(dpnodepath1)-1; pathindex++ {
			p1 := dpnodepath1[pathindex]
			p2 := dpnodepath1[pathindex+1]

			minDisSquared := 10000000000.0
			var minDisCollion *Body

			for i := 0; i < len(*staticbodys); i++ {
				//if (staticbodys[i] == dpnode->collions || staticbodys[i] == dpnode->my) {
				if (*staticbodys)[i] == dpnode.my {
					continue
				}
				R := vec2d.Add((*staticbodys)[i].R, dpnode.my.R)
				//
				if this.IsSegmentCollionSquare(p1, p2, (*staticbodys)[i].Position, R) {
					//继续绕路
					if (*staticbodys)[i] == dpnode.collions {
						log.Info("staticbodys[i] == dpnode->collions---%d", (*staticbodys)[i].Tag)
					}
					t1 := vec2d.Sub((*staticbodys)[i].Position, p1)
					disSquared := t1.LengthSquared()
					if minDisCollion == nil {
						minDisCollion = (*staticbodys)[i]
						minDisSquared = disSquared
					} else {
						if minDisSquared > disSquared {
							minDisCollion = (*staticbodys)[i]
							minDisSquared = disSquared
						}
					}

				} else {

				}
			}
			if minDisCollion != nil {
				//如果与之前的所有父节点有碰撞 则不能通行
				parent := dpnode.parent
				isbreak := false
				for {
					if parent == nil {
						break
					}
					if minDisCollion == parent.collions {
						isbreak = true
						break
					}
					parent = parent.parent
				}
				if isbreak {
					canPassAblePath1 = false
					isbreakpath = true
					break
				}
				R := vec2d.Add(minDisCollion.R, dpnode.my.R)

				path1 := make([]vec2d.Vec2, 0)
				path2 := make([]vec2d.Vec2, 0)

				detourPointIndex := pathindex + 1
				for {
					if detourPointIndex >= len(dpnodepath1) {
						canPassAblePath1 = false
						isbreakpath = true
						break
					}
					//log.Info("----DoCheckGameData--tag:%d", minDisCollion.Tag)
					//cocos2d::log("--------tag:%d", minDisCollion->Tag);

					if this.CalcDetourPathFromSquare(p1, minDisCollion.Position, R, dpnodepath1[detourPointIndex], &path1, &path2) {

						this.ChangeErrorPath(dpnode.my, minDisCollion, staticbodys, &path1, &path2)

						var dpNode1 DetourPathNode
						dpNode1.parent = dpnode
						dpNode1.collions = minDisCollion
						dpNode1.my = dpnode.my

						dpNode1.serachIndex = pathindex
						first := append([]vec2d.Vec2{}, dpnodepath1[:pathindex+1]...)
						first2 := append([]vec2d.Vec2{}, dpnodepath1[:pathindex+1]...)
						rear := append([]vec2d.Vec2{}, dpnodepath1[detourPointIndex:]...)
						rear2 := append([]vec2d.Vec2{}, dpnodepath1[detourPointIndex:]...)

						dpNode1.path1 = make([]vec2d.Vec2, 0)
						dpNode1.path1 = append(first, path1[:]...)
						dpNode1.path1 = append(dpNode1.path1, rear...)

						dpNode1.path2 = make([]vec2d.Vec2, 0)
						dpNode1.path2 = append(first2, path2[:]...)
						dpNode1.path2 = append(dpNode1.path2, rear2...)

						canPassAblePath1 = this.CheckDetourPathNode1(&dpNode1, staticbodys, path)
						if canPassAblePath1 == true {
							//log.Info("--------canPassAblePath1--------------")
							return true
						}
						isbreakpath = true
						break
					} else {
						//此目标点 不能绕行
						//canPassAblePath1 = false;

					}
					detourPointIndex++
				}

				break
			}
			if isbreakpath == true {
				break
			}

		}
		if canPassAblePath1 == true {
			(*path) = make([]vec2d.Vec2, len(dpnodepath1))
			copy((*path), dpnodepath1)
			return true
		} else {
			//return false;
		}

	}

	return false
}
func (this *WardCore) CalcDetourPath(my *Body, collion *Body, targetPos vec2d.Vec2, path *[]vec2d.Vec2) {
	(*path) = make([]vec2d.Vec2, 0)
	//目标点被当前障碍物阻碍
	R := vec2d.Add(collion.R, my.R)
	if collion.NextPosition.X-R.X < targetPos.X && collion.NextPosition.X+R.X > targetPos.X &&
		collion.NextPosition.Y-R.Y < targetPos.Y && collion.NextPosition.Y+R.Y > targetPos.Y {
		return
	}

	var path1, path2 []vec2d.Vec2
	this.CalcDetourPathFromSquare(my.Position, collion.Position, R, targetPos, &path1, &path2)

	var dpNode DetourPathNode
	dpNode.parent = nil
	dpNode.collions = collion
	dpNode.my = my
	dpNode.serachIndex = 0
	dpNode.path1 = append(dpNode.path1, my.Position)
	dpNode.path1 = append(dpNode.path1, path1[:]...)
	dpNode.path1 = append(dpNode.path1, targetPos)

	dpNode.path2 = append(dpNode.path2, my.Position)
	dpNode.path2 = append(dpNode.path2, path2[:]...)
	dpNode.path2 = append(dpNode.path2, targetPos)

	var bodys []*Body
	this.GetStaticBodys(&bodys)
	if this.CheckDetourPathNodeT(&dpNode, &bodys, path) {
		log.Info("1111111111111")
	} else {
		log.Info("2222222222222")
	}
}
func (this *WardCore) GetNextPositionCollision(one *Body) *Body {
	for i := 0; i < len(this.Bodys); i++ {
		if this.Bodys[i] != one {
			R := vec2d.Add(this.Bodys[i].R, one.R)
			if this.Bodys[i].Position.X-R.X < one.NextPosition.X && this.Bodys[i].Position.X+R.X > one.NextPosition.X &&
				this.Bodys[i].Position.Y-R.Y < one.NextPosition.Y && this.Bodys[i].Position.Y+R.Y > one.NextPosition.Y {
				return this.Bodys[i]
			}
		}
	}
	return nil
}

func (this *WardCore) Update(dt float64) {
	for i := 0; i < len(this.Bodys); i++ {
		this.Bodys[i].Update(dt)
	}
}
func (this *WardCore) CreateBody(position vec2d.Vec2, r vec2d.Vec2, speedsize float64) *Body {
	body := &Body{}
	body.Position = position
	body.R = r
	body.SpeedSize = speedsize
	body.Core = this

	this.Bodys = append(this.Bodys, body)
	return body
}

//	Body* Core::CreateBody(cocos2d::Vec2 position, cocos2d::Vec2 r, float speedsize)
//	{
//		Body* body = new Body(this);
//		body->Position = position;
//		body->R = r;
//		body->SpeedSize = speedsize;
//		//body->TargetPosition.push_back(targetPos);

//		Bodys.push_back(body);
//		return body;
//	}
