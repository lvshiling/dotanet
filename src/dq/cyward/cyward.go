package cyward

import (
	"dq/log"
	"dq/utils"
	"dq/vec2d"
	"math"
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

type MyPolygon struct {
	Center vec2d.Vec2
	Points []vec2d.Vec2 //绝对坐标 左上 右上 右下 左下
	IsRect bool
}

func multi(p1 vec2d.Vec2, p2 vec2d.Vec2, p0 vec2d.Vec2) float64 {
	return (p1.X-p0.X)*(p2.Y-p0.Y) - (p2.X-p0.X)*(p1.Y-p0.Y)
}

/*
	令矢量的起点为p1，终点为p2，判断的点为p3，
		如果fangxiang（p1，p2，p3）为正数，则C在矢量AB的左侧；
		如果fangxiang（p1，p2，p3）为负数，则C在矢量AB的右侧；
		如果fangxiang（p1，p2，p3）为0，则C在直线AB上。*/
func fangxiang(p1 vec2d.Vec2, p2 vec2d.Vec2, p3 vec2d.Vec2) float64 {
	var a = (p1.X-p3.X)*(p2.Y-p3.Y) - (p1.Y-p3.Y)*(p2.X-p3.X)
	return a
}

func (this *MyPolygon) IsInMyPolygon(p vec2d.Vec2) bool {

	if len(this.Points) < 4 {
		return true
	}
	if this.IsRect {
		if this.Points[0].X <= p.X && this.Points[1].X >= p.X && this.Points[3].Y <= p.Y && this.Points[0].Y >= p.Y {
			return true
		}
		return false
	} else {
		var pre, now float64
		var n = len(this.Points)
		for i := 0; i < n; i++ {
			now = multi(p, this.Points[i], this.Points[(i+1)%n])
			if i > 0 {
				if pre*now < 0 {
					return false
				}
			}
			pre = now
		}
		return true
	}
}

func (this *MyPolygon) GetBigMyPolygonOnePoint(pointindex int, r vec2d.Vec2) vec2d.Vec2 {
	if this.IsRect {
		var offset [4]vec2d.Vec2
		offset[0] = vec2d.Vec2{-r.X, r.Y}
		offset[1] = vec2d.Vec2{r.X, r.Y}
		offset[2] = vec2d.Vec2{r.X, -r.Y}
		offset[3] = vec2d.Vec2{-r.X, -r.Y}

		re := vec2d.Add(this.Points[pointindex], offset[pointindex])
		return re

	} else {
		var addlen = r.Length()
		var dir = vec2d.Sub(this.Points[pointindex], this.Center)
		add := vec2d.Add(this.Points[pointindex], vec2d.Mul(dir.GetNormalized(), addlen))

		return add

	}
}

type Body struct {
	Core           *WardCore
	Position       vec2d.Vec2 //当前位置
	R              vec2d.Vec2 //矩形半径
	SpeedSize      float64    //移动速度大小
	CollisoinLevel int32      // 碰撞等级 障碍物石头为2 普通单位为1 只计算大于等于自己碰撞等级的body的碰撞
	IsCollisoin    bool       //是否计算碰撞

	MoveDir        vec2d.Vec2   //移动目标方向
	TargetPosition []vec2d.Vec2 //移动目标位置
	DetourPath     []vec2d.Vec2 //绕路 路径

	CollisoinStopTime float64 //碰撞停止移动剩余时间
	CurSpeedSize      float64 //当前速度大小

	TargetIndex   int        //计算后的目标位置索引
	NextPosition  vec2d.Vec2 //计算后的下一帧位置
	Direction     vec2d.Vec2 //速度方向
	TurnDirection bool       //是否改变朝向

	Tag int //标记

	IsRect       bool         //是否是标准矩形
	OffsetPoints []vec2d.Vec2 //相对偏移位置 坐标 左上 右上 右下 左下
	M_MyPolygon  *MyPolygon
}

func (this *Body) GetMyPolygonBig(p1 *Body, big vec2d.Vec2) *MyPolygon {
	p1r := vec2d.Vec2{}
	p1r.X = 0
	p1r.Y = 0
	if p1 != nil {
		p1r = p1.R
	}

	this.M_MyPolygon.Points = this.M_MyPolygon.Points[0:0]
	this.M_MyPolygon.IsRect = this.IsRect
	this.M_MyPolygon.Center = this.Position
	if this.IsRect {

		r := vec2d.Add(vec2d.Add(p1r, this.R), big)
		//log.Info("GetMyPolygonBig:r :%v  %v  %v  %v", r, p1r, this.R, big)

		//变成正方形
		this.M_MyPolygon.Points = append(this.M_MyPolygon.Points, vec2d.Add(this.M_MyPolygon.Center, vec2d.Vec2{-r.X, r.Y}))
		this.M_MyPolygon.Points = append(this.M_MyPolygon.Points, vec2d.Add(this.M_MyPolygon.Center, vec2d.Vec2{r.X, r.Y}))
		this.M_MyPolygon.Points = append(this.M_MyPolygon.Points, vec2d.Add(this.M_MyPolygon.Center, vec2d.Vec2{r.X, -r.Y}))
		this.M_MyPolygon.Points = append(this.M_MyPolygon.Points, vec2d.Add(this.M_MyPolygon.Center, vec2d.Vec2{-r.X, -r.Y}))

	} else {
		tt := vec2d.Add(p1r, big)
		addlen := tt.Length()
		for i := 0; i < len(this.OffsetPoints); i++ {
			add := vec2d.Add(vec2d.Add(this.M_MyPolygon.Center, this.OffsetPoints[i]), vec2d.Mul(this.OffsetPoints[i].GetNormalized(), addlen))
			this.M_MyPolygon.Points = append(this.M_MyPolygon.Points, add)
			//this.M_MyPolygon->m_Points.push_back(this.M_MyPolygon->m_Center + m_OffsetPoints[i] + ((m_OffsetPoints[i]).getNormalized() * addlen));
		}
	}

	return this.M_MyPolygon
}
func (this *Body) GetMyPolygon(p1 *Body) *MyPolygon {
	return this.GetMyPolygonBig(p1, vec2d.Vec2{0, 0})
}

func (this *Body) SetTag(tag int) {
	this.Tag = tag
}
func (this *Body) Update(dt float64) {
	//log.Info(" %v---%v", dt, this.CollisoinStopTime)
	this.CollisoinStopTime -= dt
	oldDirection := this.Direction
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

	if this.TurnDirection == false {
		this.Direction = oldDirection
	}

}
func (this *Body) ClearMoveDirAndMoveTarget() {
	//清空目的地
	this.TargetPosition = this.TargetPosition[0:0]
	this.DetourPath = this.DetourPath[0:0]
	this.CollisoinStopTime = 0
	this.MoveDir = vec2d.Vec2{}
}

//设置移动方向
func (this *Body) SetMoveDir(dir vec2d.Vec2) {
	//清空目的地
	this.TargetPosition = this.TargetPosition[0:0]
	this.DetourPath = this.DetourPath[0:0]
	this.CollisoinStopTime = 0
	this.MoveDir = dir
}

//瞬间移动到目的地
func (this *Body) BlinkToPos(pos vec2d.Vec2, rotate float64) {
	//log.Info("BlinkToPos")
	//清空以前的目的地
	this.ClearMoveDirAndMoveTarget()
	collisoin := this.Core.GetTargetPosCollision(this, pos)
	//mypolygon2 := nil
	intersectPoint := pos
	count1 := 0
	if collisoin != nil {

		calcing := make([]interface{}, 0)
		calced := make([]interface{}, 0)
		calcing = append(calcing, collisoin)
		for {
			if len(calcing) <= 0 {
				break
			}
			if count1 >= 1000 {
				log.Info("BlinkToPos >= 1000 count %f  %f", pos.X, pos.Y)
				break
			}
			collisoin = calcing[0].(*Body)
			//log.Info("tag000=====%d ", collisoin.Tag)

			mypolygon2 := collisoin.GetMyPolygonBig(this, vec2d.Vec2{X: 0.01, Y: 0.01})
			//mypolygon3 := collisoin.GetMyPolygon(this)
			//log.Info("111:%v %v %v ", mypolygon2, this.Position, pos)

			dir2 := vec2d.Sub(this.Position, collisoin.Position)
			dir2.Rotate(rotate)
			dir2.Normalize()
			dir2.MulToFloat64(10000.0)
			dir2.Collect(&collisoin.Position)

			if this.Core.GetSegmentInsterset(dir2, collisoin.Position, mypolygon2, &intersectPoint) {
				//log.Info("222:%v ", mypolygon2)
				collisoin1 := this.Core.GetTargetPosCollision(this, intersectPoint)
				//log.Info("333:%v ", mypolygon2)
				if collisoin1 != nil {

					if utils.FindFromSlice(calcing, collisoin1) == nil && utils.FindFromSlice(calced, collisoin1) == nil {
						calcing = append(calcing, collisoin1)
					}

					count := 20.0
					for i := 1.0; i <= count; i++ {
						//Rotate
						dir := vec2d.Sub(this.Position, collisoin.Position)
						dir.Rotate(float64(i*(180.0/count)) + rotate)
						dir.Normalize()
						dir.MulToFloat64(10000.0)
						dir.Collect(&collisoin.Position)
						//log.Info("111=====%f %f", dir.X, dir.Y)
						mypolygon2 = collisoin.GetMyPolygonBig(this, vec2d.Vec2{X: 0.01, Y: 0.01})
						if this.Core.GetSegmentInsterset(dir, collisoin.Position, mypolygon2, &intersectPoint) {
							collisoin1 = this.Core.GetTargetPosCollision(this, intersectPoint)
							if collisoin1 == nil {
								//设置真实值
								this.Position = intersectPoint
								return
							} else {
								//log.Info("tag111=====%d  %f  %v  %v", collisoin1.Tag, i, mypolygon2, intersectPoint)
								if utils.FindFromSlice(calcing, collisoin1) == nil && utils.FindFromSlice(calced, collisoin1) == nil {
									calcing = append(calcing, collisoin1)
								}
							}
						}

						dir2 := vec2d.Sub(this.Position, collisoin.Position)
						dir2.Rotate(float64(0.0-i*(180.0/count)) + rotate)
						dir2.Normalize()
						dir2.MulToFloat64(10000.0)
						dir2.Collect(&collisoin.Position)
						//log.Info("222=====%f %f", dir2.X, dir2.Y)
						mypolygon2 = collisoin.GetMyPolygonBig(this, vec2d.Vec2{X: 0.01, Y: 0.01})
						if this.Core.GetSegmentInsterset(dir2, collisoin.Position, mypolygon2, &intersectPoint) {
							collisoin1 = this.Core.GetTargetPosCollision(this, intersectPoint)
							if collisoin1 == nil {
								//设置真实值
								this.Position = intersectPoint
								return
							} else {
								//log.Info("tag222=====%d  %f  %v  %v", collisoin1.Tag, i, mypolygon2, intersectPoint)
								if utils.FindFromSlice(calcing, collisoin1) == nil && utils.FindFromSlice(calced, collisoin1) == nil {
									calcing = append(calcing, collisoin1)
								}
							}
						}
					}
				}

				//从搜索表中删除本body  加入已经搜索表
				calced = append(calced, collisoin)
				calcing = append(calcing[:0], calcing[1:]...)
				//log.Info("len====%d  %d", len(calcing), len(calced))
			} else {

			}
			//
			count1++
		}

	}
	//设置真实值
	this.Position = intersectPoint

}

//设置目的地
func (this *Body) SetTarget(pos vec2d.Vec2) {

	//log.Info("SetTarget %f  %f", pos.X, pos.Y)
	if len(this.TargetPosition) > 0 {
		//
		t1 := vec2d.Sub(this.TargetPosition[len(this.TargetPosition)-1], pos)
		if t1.LengthSquared() <= 0.0001 {
			//log.Info("SetTarget repeat!!")
			return
		}
	}
	this.ClearMoveDirAndMoveTarget()
	//this.TargetPosition = this.TargetPosition[0:0]
	//this.DetourPath = this.DetourPath[0:0]
	this.TargetPosition = append(this.TargetPosition, pos)
	//this.MoveDir = vec2d.Vec2{}
	//this.CollisoinStopTime = 0

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
	this.Core.GetStaticBodysNoTarget(&bodys, this.Position, pos, this)
	if this.Core.CheckDetourPathNodeT(dpNode, &bodys, &this.DetourPath) {
		//log.Info("SetTarget %d", this.Tag)
		//		for i := 0; i < len(this.DetourPath); i++ {
		//			log.Info("x: %f  y:%f", this.DetourPath[i].X, this.DetourPath[i].Y)
		//		}
	} else {
		//log.Info("SetTarget 222")
	}

	t2 := time.Now().UnixNano()
	if (t2-t1)/1e6 > 10 {
		log.Info("time:%d", (t2-t1)/1e6)
	}

}

//如果已经在碰撞区域内  就可以直接走出来，不需要再检查碰撞
func (this *Body) CheckPositionCollisoin(dt float64) *Body {

	if this.Core.GetCurPositionCollision(this) == nil {
		return this.Core.GetNextPositionCollision(this)
	}

	return nil
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

func (this *Body) IsMove() bool {

	//log.Info("IsMove:%d---%d---%f", len(this.TargetPosition), len(this.DetourPath), this.CollisoinStopTime)
	if (len(this.TargetPosition) <= 0 && len(this.DetourPath) <= 0) || this.CollisoinStopTime > 0.0000001 {

		return false
	}
	return true
}

func (this *Body) CalcNextPosition(dt float64) bool {

	//移动方向
	if this.MoveDir.IsEqual(&vec2d.Vec2{}) == false {
		startpos := this.Position
		movedis := this.SpeedSize * dt

		this.Direction = this.MoveDir.GetNormalized()
		this.NextPosition = vec2d.Add(startpos, vec2d.Mul(this.Direction, movedis))

		collisionOne := this.CheckPositionCollisoin(0)
		if collisionOne != nil {

			mypolygon := collisionOne.GetMyPolygon(this)
			var points2P1Index []int
			this.Core.GetPointIndexFromSquare(mypolygon, this.Position, &points2P1Index)
			//删掉中间点(只保留两端顶点)
			if len(points2P1Index) >= 3 {
				points2P1Index = append(points2P1Index[:1], points2P1Index[2:]...)
			}
			//获取2可通行的两点
			point1 := mypolygon.GetBigMyPolygonOnePoint(points2P1Index[0], vec2d.Vec2{0.01, 0.01})
			point2 := mypolygon.GetBigMyPolygonOnePoint(points2P1Index[1], vec2d.Vec2{0.01, 0.01})
			d1 := vec2d.Sub(point1, this.Position)
			d2 := vec2d.Sub(point2, this.Position)
			a1 := vec2d.Angle(d1, this.Direction)
			a2 := vec2d.Angle(d2, this.Direction)
			if math.Abs(a1) < math.Abs(a2) {
				this.Direction = d1
			} else {
				this.Direction = d2
			}
			this.NextPosition = vec2d.Add(startpos, vec2d.Mul(this.Direction.GetNormalized(), movedis))

			collisionOne := this.CheckPositionCollisoin(0)
			if collisionOne != nil {
				return false
			}

		}

		//设置最终值
		this.Position = this.NextPosition
		this.CurSpeedSize = this.SpeedSize

		return false
	} else { //移动目的地
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

}

//线段是否与矩形相交
func (this *WardCore) IsSegmentCollionSquare(p1 vec2d.Vec2, p2 vec2d.Vec2, mypolygon *MyPolygon) bool {
	//变成正方形
	circlep1 := mypolygon.Points[0]
	circlep2 := mypolygon.Points[1]
	circlep3 := mypolygon.Points[2]
	circlep4 := mypolygon.Points[3]

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
func (this *WardCore) GetSegmentInsterset(p1 vec2d.Vec2, p2 vec2d.Vec2, mypolygon *MyPolygon, Re *vec2d.Vec2) bool {
	//变成正方形
	circlep1 := mypolygon.Points[0]
	circlep2 := mypolygon.Points[1]
	circlep3 := mypolygon.Points[2]
	circlep4 := mypolygon.Points[3]

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
		//log.Info("GetSegmentInsterset:%v %v %v %v    %v  %v", circlep1, circlep2, circlep3, circlep4, p1, p2)
		return false
	}
}

func (this *WardCore) GetPointIndexFromSquare(mypolygon *MyPolygon, targetPos vec2d.Vec2, posIndex *[]int) {
	//正方形的4个顶点
	if mypolygon.IsRect {
		if targetPos.X <= mypolygon.Points[0].X && targetPos.Y >= mypolygon.Points[0].Y { //目标点在矩形的左上
			(*posIndex) = append((*posIndex), 3)
			(*posIndex) = append((*posIndex), 0)
			(*posIndex) = append((*posIndex), 1)
		} else if targetPos.X >= mypolygon.Points[0].X && targetPos.X <= mypolygon.Points[1].X && targetPos.Y >= mypolygon.Points[0].Y { //正上
			(*posIndex) = append((*posIndex), 0)
			(*posIndex) = append((*posIndex), 1)
		} else if targetPos.X >= mypolygon.Points[1].X && targetPos.Y >= mypolygon.Points[0].Y { //右上
			(*posIndex) = append((*posIndex), 0)
			(*posIndex) = append((*posIndex), 1)
			(*posIndex) = append((*posIndex), 2)
		} else if targetPos.X >= mypolygon.Points[1].X && targetPos.Y < mypolygon.Points[1].Y && targetPos.Y >= mypolygon.Points[2].Y { //正右
			(*posIndex) = append((*posIndex), 1)
			(*posIndex) = append((*posIndex), 2)
		} else if targetPos.X >= mypolygon.Points[1].X && targetPos.Y <= mypolygon.Points[2].Y { //右下
			(*posIndex) = append((*posIndex), 1)
			(*posIndex) = append((*posIndex), 2)
			(*posIndex) = append((*posIndex), 3)
		} else if targetPos.X >= mypolygon.Points[3].X && targetPos.X <= mypolygon.Points[2].X && targetPos.Y <= mypolygon.Points[2].Y { //正下
			(*posIndex) = append((*posIndex), 2)
			(*posIndex) = append((*posIndex), 3)
		} else if targetPos.X <= mypolygon.Points[3].X && targetPos.Y <= mypolygon.Points[2].Y { //左下
			(*posIndex) = append((*posIndex), 2)
			(*posIndex) = append((*posIndex), 3)
			(*posIndex) = append((*posIndex), 0)
		} else if targetPos.X <= mypolygon.Points[3].X && targetPos.Y <= mypolygon.Points[0].Y && targetPos.Y >= mypolygon.Points[3].Y { //正左
			(*posIndex) = append((*posIndex), 3)
			(*posIndex) = append((*posIndex), 0)
		}
	} else {
		var isLeft1 = false
		if fangxiang(mypolygon.Points[0], mypolygon.Points[1], targetPos) >= 0 {
			isLeft1 = true
		}
		var isLeft2 = false
		if fangxiang(mypolygon.Points[1], mypolygon.Points[2], targetPos) >= 0 {
			isLeft2 = true
		}
		var isLeft3 = false
		if fangxiang(mypolygon.Points[2], mypolygon.Points[3], targetPos) >= 0 {
			isLeft3 = true
		}
		var isLeft4 = false
		if fangxiang(mypolygon.Points[3], mypolygon.Points[0], targetPos) >= 0 {
			isLeft4 = true
		}
		if isLeft1 && isLeft4 { //目标点在矩形的左上
			(*posIndex) = append((*posIndex), 3)
			(*posIndex) = append((*posIndex), 0)
			(*posIndex) = append((*posIndex), 1)
		} else if isLeft1 && !isLeft4 && !isLeft2 { //正上
			(*posIndex) = append((*posIndex), 0)
			(*posIndex) = append((*posIndex), 1)
		} else if isLeft1 && isLeft2 { //右上
			(*posIndex) = append((*posIndex), 0)
			(*posIndex) = append((*posIndex), 1)
			(*posIndex) = append((*posIndex), 2)
		} else if isLeft2 && !isLeft1 && !isLeft3 { //正右
			(*posIndex) = append((*posIndex), 1)
			(*posIndex) = append((*posIndex), 2)
		} else if isLeft2 && isLeft3 { //右下
			(*posIndex) = append((*posIndex), 1)
			(*posIndex) = append((*posIndex), 2)
			(*posIndex) = append((*posIndex), 3)
		} else if isLeft3 && !isLeft2 && !isLeft4 { //正下
			(*posIndex) = append((*posIndex), 2)
			(*posIndex) = append((*posIndex), 3)
		} else if isLeft3 && isLeft4 { //左下
			(*posIndex) = append((*posIndex), 2)
			(*posIndex) = append((*posIndex), 3)
			(*posIndex) = append((*posIndex), 0)
		} else if isLeft4 && !isLeft3 && !isLeft1 { //正左
			(*posIndex) = append((*posIndex), 3)
			(*posIndex) = append((*posIndex), 0)
		}
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
func (this *WardCore) CalcDetourPathFromSquare(p1 vec2d.Vec2, mypolygon *MyPolygon, targetPos vec2d.Vec2, path1 *[]vec2d.Vec2, path2 *[]vec2d.Vec2) bool {
	//目标点在正方形内部
	if mypolygon.IsInMyPolygon(targetPos) {
		return false
	}

	if mypolygon.IsInMyPolygon(p1) {
		return false
	}
	//r = r + 1;
	//正方形的4个顶点
	//计算目标点能直接通过的顶点
	var points2TargetIndex []int
	var points2P1Index []int

	this.GetPointIndexFromSquare(mypolygon, targetPos, &points2TargetIndex)
	this.GetPointIndexFromSquare(mypolygon, p1, &points2P1Index)
	//删掉中间点(只保留两端顶点)
	if len(points2P1Index) >= 3 {
		points2P1Index = append(points2P1Index[:1], points2P1Index[2:]...)
	}
	rightp := points2P1Index[0]
	leftp := points2P1Index[1]

	//向外偏移一个像素
	has := false
	for {
		if has == true {
			break
		}
		(*path1) = append((*path1), mypolygon.GetBigMyPolygonOnePoint(rightp, vec2d.Vec2{0.01, 0.01}))

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
		(*path2) = append((*path2), mypolygon.GetBigMyPolygonOnePoint(leftp, vec2d.Vec2{0.01, 0.01}))
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
			//R := vec2d.Add(my.R, (*staticbodys)[i].R)
			mypolygon1 := (*staticbodys)[i].GetMyPolygon(my)
			if mypolygon1.IsInMyPolygon((*path1)[j]) {
				//更改点
				dir := vec2d.Sub((*path1)[j], detourBody.Position)
				seg := vec2d.Add(vec2d.Mul(dir.GetNormalized(), 10000), (*path1)[j])

				var intersectPoint vec2d.Vec2
				mypolygon2 := (*staticbodys)[i].GetMyPolygonBig(my, vec2d.Vec2{0.01, 0.01})
				if this.GetSegmentInsterset((*path1)[j], seg, mypolygon2, &intersectPoint) {
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
			//R := vec2d.Add(my.R, (*staticbodys)[i].R)
			mypolygon1 := (*staticbodys)[i].GetMyPolygon(my)
			if mypolygon1.IsInMyPolygon((*path2)[j]) {
				//更改点
				dir := vec2d.Sub((*path2)[j], detourBody.Position)
				seg := vec2d.Add(vec2d.Mul(dir.GetNormalized(), 10000), (*path2)[j])

				var intersectPoint vec2d.Vec2
				mypolygon2 := (*staticbodys)[i].GetMyPolygonBig(my, vec2d.Vec2{0.01, 0.01})
				if this.GetSegmentInsterset((*path2)[j], seg, mypolygon2, &intersectPoint) {
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
				//R := vec2d.Add((*staticbodys)[i].R, me.R)
				//
				mypolygon := (*staticbodys)[i].GetMyPolygon(me)
				if this.IsSegmentCollionSquare(p1, p2, mypolygon) {
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
				//R := vec2d.Add((*staticbodys)[i].R, dpnode.my.R)
				//
				mypolygon := (*staticbodys)[i].GetMyPolygon(dpnode.my)
				if this.IsSegmentCollionSquare(p1, p2, mypolygon) {
					//继续绕路
					if (*staticbodys)[i] == dpnode.collions {
						//log.Info("staticbodys[i] == dpnode->collions---%d", (*staticbodys)[i].Tag)
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
				//R := vec2d.Add(minDisCollion.R, dpnode.my.R)

				path1 := make([]vec2d.Vec2, 0)
				path2 := make([]vec2d.Vec2, 0)

				detourPointIndex := pathindex + 1
				mypolygon1 := minDisCollion.GetMyPolygon(dpnode.my)
				for {
					if detourPointIndex >= len(dpnodepath1) {
						canPassAblePath1 = false
						isbreakpath = true
						break
					}
					//log.Info("----DoCheckGameData--tag:%d", minDisCollion.Tag)
					//cocos2d::log("--------tag:%d", minDisCollion->Tag);

					if this.CalcDetourPathFromSquare(p1, mypolygon1, dpnodepath1[detourPointIndex], &path1, &path2) {

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
				//R := vec2d.Add((*staticbodys)[i].R, dpnode.my.R)
				//-----------区别
				mypolygon := (*staticbodys)[i].GetMyPolygon(dpnode.my)
				if this.IsSegmentCollionSquare(p1, p2, mypolygon) {
					//继续绕路
					if (*staticbodys)[i] == dpnode.collions {
						//log.Info("staticbodys[i] == dpnode->collions---%d", (*staticbodys)[i].Tag)
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
				//R := vec2d.Add(minDisCollion.R, dpnode.my.R)

				path1 := make([]vec2d.Vec2, 0)
				path2 := make([]vec2d.Vec2, 0)

				detourPointIndex := pathindex + 1
				mypolygon1 := minDisCollion.GetMyPolygon(dpnode.my)
				for {
					if detourPointIndex >= len(dpnodepath1) {
						canPassAblePath1 = false
						isbreakpath = true
						break
					}
					//log.Info("----DoCheckGameData--tag:%d", minDisCollion.Tag)
					//cocos2d::log("--------tag:%d", minDisCollion->Tag);
					//-----------区别
					if this.CalcDetourPathFromSquare(p1, mypolygon1, dpnodepath1[detourPointIndex], &path1, &path2) {

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
	//R := vec2d.Add(collion.R, my.R)
	mypolygon1 := collion.GetMyPolygon(my)
	if mypolygon1.IsInMyPolygon(targetPos) {
		return
	}

	var path1, path2 []vec2d.Vec2
	this.CalcDetourPathFromSquare(my.Position, mypolygon1, targetPos, &path1, &path2)

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
	this.GetStaticBodysNoTarget(&bodys, my.Position, targetPos, my)
	if this.CheckDetourPathNodeT(&dpNode, &bodys, path) {
		//log.Info("1111111111111")
	} else {
		//log.Info("2222222222222")
	}
}

var gWardCore_ZoneWidth float64 = 4.0
var gWardCore_ZoneHeight float64 = 4.0

func CheckCalcCollisoin(v *Body, my *Body) bool {
	if v == my {
		return false
	}
	if v.IsCollisoin == false {
		return false
	}
	if v.CollisoinLevel <= my.CollisoinLevel && my.IsCollisoin == false {

		return false

	}
	return true
}

//避障核心
type WardCore struct {
	Bodys map[*Body]*Body
	//Bodys []*Body
	ZoneBodys map[utils.SceneZone][]*Body
}

//获取静止且不包含目的点 的body
func (this *WardCore) GetStaticBodysNoTarget(bodys *[]*Body, mypos vec2d.Vec2, target vec2d.Vec2, mybody *Body) {

	zones := utils.GetVisibleZonesFromWH_Two((mypos.X), (mypos.Y), target.X, target.Y, gWardCore_ZoneWidth, gWardCore_ZoneHeight)
	//遍历可视区域
	for _, vzone := range zones {
		if _, ok := this.ZoneBodys[vzone]; ok {
			//遍历区域中的单位
			for _, v := range this.ZoneBodys[vzone] {
				if v.CurSpeedSize <= 0 && CheckCalcCollisoin(v, mybody) {
					mypolygon1 := v.GetMyPolygon(nil)
					if mypolygon1.IsInMyPolygon(target) {
						continue
					}
					(*bodys) = append((*bodys), v)
				}
			}
		}
	}

	//	for _, v := range this.Bodys {
	//		if v.CurSpeedSize <= 0 {
	//			mypolygon1 := v.GetMyPolygon(nil)
	//			if mypolygon1.IsInMyPolygon(target) {
	//				continue
	//			}
	//			(*bodys) = append((*bodys), v)
	//		}
	//	}
}

//获取静止的body
//func (this *WardCore) GetStaticBodys(bodys *[]*Body) {

//	for _, v := range this.Bodys {
//		if v.CurSpeedSize <= 0 {
//			(*bodys) = append((*bodys), v)
//		}
//	}
//}

func (this *WardCore) GetCurPositionCollision(one *Body) *Body {

	for _, v := range this.Bodys {
		if CheckCalcCollisoin(v, one) {

			mypolygon1 := v.GetMyPolygon(one)
			if mypolygon1.IsInMyPolygon(one.Position) {
				return v
			}
		}
	}

	return nil
}

func (this *WardCore) GetNextPositionCollision(one *Body) *Body {

	for _, v := range this.Bodys {
		if CheckCalcCollisoin(v, one) {

			mypolygon1 := v.GetMyPolygon(one)
			if mypolygon1.IsInMyPolygon(one.NextPosition) {
				return v
			}
		}
	}

	return nil
}

//获取目标位置的碰撞体
func (this *WardCore) GetTargetPosCollision(one *Body, targetpos vec2d.Vec2) *Body {

	for _, v := range this.Bodys {
		if CheckCalcCollisoin(v, one) {

			mypolygon1 := v.GetMyPolygon(one)
			if mypolygon1.IsInMyPolygon(targetpos) {
				return v
			}
		}
	}

	return nil
}

//处理分区
func (this *WardCore) DoZone() {
	this.ZoneBodys = make(map[utils.SceneZone][]*Body)
	for _, v := range this.Bodys {

		zone := utils.GetSceneZoneFromWH((v.Position.X), (v.Position.Y), gWardCore_ZoneWidth, gWardCore_ZoneHeight)
		this.ZoneBodys[zone] = append(this.ZoneBodys[zone], v)

	}
}

func (this *WardCore) Update(dt float64) {

	this.DoZone()
	for _, v := range this.Bodys {
		v.Update(dt)
	}
}
func (this *WardCore) CreateBody(position vec2d.Vec2, r vec2d.Vec2, speedsize float64, level int32) *Body {
	body := &Body{}
	body.Position = position
	body.R = r
	body.SpeedSize = speedsize
	body.Core = this
	body.IsRect = true
	body.M_MyPolygon = &MyPolygon{}
	body.MoveDir = vec2d.Vec2{X: 0, Y: 0}
	//CollisoinLevel int32      // 碰撞等级 障碍物石头为2 普通单位为1
	body.CollisoinLevel = level
	body.IsCollisoin = true
	body.TurnDirection = true
	this.Bodys[body] = body

	//body.BlinkToPos(position)
	return body
}
func (this *WardCore) CreateBodyPolygon(position vec2d.Vec2, points []vec2d.Vec2, speedsize float64, level int32) *Body {
	body := &Body{}
	body.Position = position
	body.SpeedSize = speedsize
	body.Core = this
	body.IsRect = false
	body.M_MyPolygon = &MyPolygon{}
	body.OffsetPoints = points
	body.MoveDir = vec2d.Vec2{X: 0, Y: 0}
	body.CollisoinLevel = level
	body.IsCollisoin = true
	body.TurnDirection = true
	this.Bodys[body] = body
	return body
}

func (this *WardCore) RemoveBody(body *Body) {

	delete(this.Bodys, body)
}

func CreateWardCore() *WardCore {
	re := &WardCore{}
	re.Bodys = make(map[*Body]*Body)
	re.ZoneBodys = make(map[utils.SceneZone][]*Body)

	return re
}
