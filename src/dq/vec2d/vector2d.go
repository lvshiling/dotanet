package vec2d

import "math"

// Vector type defines vector using exported float64 values: X and Y
type Vec2 struct {
	X, Y float64
}

// Init initializes already created vector
func (v *Vec2) Init(x, y float64) *Vec2 {
	v.X = x
	v.Y = y
	return v
}

// New returns a new vector
func New(x, y float64) *Vec2 {
	return new(Vec2).Init(x, y)
}

// IsEqual compares а Vector with another and returns true if they're equal
func (v *Vec2) IsEqual(other *Vec2) bool {
	return v.X == other.X && v.Y == other.Y
}

// Angle returns the Vector's angle in float64
func (v *Vec2) Angle() float64 {
	return math.Atan2(v.Y, v.X) / (math.Pi / 180)
}

// SetAngle changes Vector's angle using vector rotation
func (v *Vec2) SetAngle(angle_degrees float64) {
	v.X = v.Length()
	v.Y = 0.0
	v.Rotate(angle_degrees)
}

// Length returns... well the Vector's length
func (v *Vec2) Length() float64 {
	return math.Sqrt(math.Pow(v.X, 2) + math.Pow(v.Y, 2))
}

func (v *Vec2) LengthSquared() float64 {
	return math.Pow(v.X, 2) + math.Pow(v.Y, 2)
}

// SetLength changes Vector's length, which obviously changes
// the values of Vector.X and Vector.Y
func (v *Vec2) SetLength(value float64) {
	length := v.Length()
	v.X *= value / length
	v.Y *= value / length
}

// Rotate Vector by given angle degrees in float64
func (v *Vec2) Rotate(angle_degrees float64) {
	radians := (math.Pi / 180) * angle_degrees
	sin := math.Sin(radians)
	cos := math.Cos(radians)

	x := v.X*cos - v.Y*sin
	y := v.X*sin + v.Y*cos
	v.X = x
	v.Y = y
}

// Collect changes Vector's X and Y by collecting them with other's
func (v *Vec2) Collect(other *Vec2) {
	v.X += other.X
	v.Y += other.Y
}

// CollectToFloat64 changes Vector's X and Y by collecting them with value
func (v *Vec2) CollectToFloat64(value float64) {
	v.X += value
	v.Y += value
}

// Sub changes Vector's X and Y by substracting them with other's
func (v *Vec2) Sub(other *Vec2) {
	v.X -= other.X
	v.Y -= other.Y
}

// SubToFloat64 changes Vector's X and Y by substracting them with value
func (v *Vec2) SubToFloat64(value float64) {
	v.X -= value
	v.Y -= value
}

// Mul changes Vector's X and Y by multiplying them with other's
func (v *Vec2) Mul(other *Vec2) {
	v.X *= other.X
	v.Y *= other.Y
}

// MulToFloat64 changes Vector's X and Y by multiplying them with value
func (v *Vec2) MulToFloat64(value float64) {
	v.X *= value
	v.Y *= value
}

// Div changes Vector's X and Y by dividing them with other's
func (v *Vec2) Div(other *Vec2) {
	v.X /= other.X
	v.Y /= other.Y
}

// DivToFloat64 changes Vector's X and Y by dividing them with value
func (v *Vec2) DivToFloat64(value float64) {
	v.X /= value
	v.Y /= value
}

func (v *Vec2) Normalize() {
	n := v.X*v.X + v.Y*v.Y
	// Already normalized.
	if n == 1.0 {
		return
	}

	n = math.Sqrt(n)
	// Too close to zero.
	if n < 0.000000000001 {
		return
	}

	n = 1.0 / n
	v.X *= n
	v.Y *= n
}

func (v *Vec2) GetNormalized() Vec2 {

	v1 := *v
	v1.Normalize()
	return v1
}

//新增
func Sub(v1 Vec2, v2 Vec2) Vec2 {
	re := Vec2{}
	re.Init(v1.X-v2.X, v1.Y-v2.Y)
	return re
}
func Add(v1 Vec2, v2 Vec2) Vec2 {
	re := Vec2{}
	re.Init(v1.X+v2.X, v1.Y+v2.Y)
	return re
}
func Mul(v1 Vec2, mul float64) Vec2 {
	re := Vec2{}
	re.Init(v1.X*mul, v1.Y*mul)
	return re
}
func Distanse(v1 Vec2, v2 Vec2) float64 {
	s := Sub(v1, v2)
	return s.Length()
}

//float Vec2::angle(const Vec2& v1, const Vec2& v2)
//{
//    float dz = v1.x * v2.y - v1.y * v2.x;
//    return atan2f(fabsf(dz) + MATH_FLOAT_SMALL, dot(v1, v2));
//}

//float Vec2::dot(const Vec2& v1, const Vec2& v2)
//{
//    return (v1.x * v2.x + v1.y * v2.y);
//}
func Dot(v1 Vec2, v2 Vec2) float64 {
	return (v1.X*v2.X + v1.Y*v2.Y)
}

//2个向量的夹角  弧度 >= 0
func Angle(v1 Vec2, v2 Vec2) float64 {
	dz := v1.X*v2.Y - v1.Y*v2.X

	return math.Atan2(math.Abs(dz)+math.SmallestNonzeroFloat32, Dot(v1, v2))
}

func CrossProduct2Vector(A Vec2, B Vec2, C Vec2, D Vec2) float64 {
	return (D.Y-C.Y)*(B.X-A.X) - (D.X-C.X)*(B.Y-A.Y)
}
func IsLineIntersect(A Vec2, B Vec2, C Vec2, D Vec2, S *float64, T *float64) bool {
	if (A.X == B.X && A.Y == B.Y) || (C.X == D.X && C.Y == D.Y) {
		return false
	}

	denom := CrossProduct2Vector(A, B, C, D)

	if denom == 0 {
		// Lines parallel or overlap
		return false
	}

	*S = CrossProduct2Vector(C, D, C, A) / denom
	*T = CrossProduct2Vector(A, B, C, A) / denom

	return true
}
func IsSegmentIntersect(A Vec2, B Vec2, C Vec2, D Vec2) bool {
	var S, T float64

	if IsLineIntersect(A, B, C, D, &S, &T) && (S >= 0.0 && S <= 1.0 && T >= 0.0 && T <= 1.0) {
		return true
	}

	return false
}
