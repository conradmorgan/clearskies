package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
)

type AlignmentPoint struct {
	ObjectId string
	Point    Vector
}

type Object struct {
	MainId    string
	Magnitude float64
	RADec     RADec
}

type RADec struct {
	RA  float64
	Dec float64
}

type CoordinateSystem struct {
	AspectRatio     float64
	AlignmentPoints [2]AlignmentPoint
	Objects         [2]Object
	IsMirrored      bool
}

type Tag struct {
	Name      string
	Magnitude float64
	Dim       float64
	Point     Vector
}

type Vector struct {
	X, Y float64
}

func (raDec RADec) ToVector() Vector {
	return Vector{raDec.RA, raDec.Dec}
}

func (v Vector) ToRADec() RADec {
	return RADec{v.X, v.Y}
}

func (v Vector) Length() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (u Vector) Dot(v Vector) float64 {
	return u.X*v.X + u.Y*v.Y
}

func (v Vector) Normalized() Vector {
	length := v.Length()
	return Vector{v.X / length, v.Y / length}
}

func (v Vector) Times(s float64) Vector {
	return Vector{v.X * s, v.Y * s}
}

func (u Vector) Sub(v Vector) Vector {
	return Vector{u.X - v.X, u.Y - v.Y}
}

func (u Vector) Plus(v Vector) Vector {
	return Vector{u.X + v.X, u.Y + v.Y}
}

func (c CoordinateSystem) ImagePointToRADec(p Vector) RADec {
	v := p.Sub(c.AlignmentPoints[0].Point)
	imgAxis1 := c.AlignmentPoints[1].Point.Sub(c.AlignmentPoints[0].Point)
	imgAxis1 = imgAxis1.Times(1 / imgAxis1.Dot(imgAxis1))
	imgAxis2 := Vector{-imgAxis1.Y, imgAxis1.X}
	raDecAxis1 := c.Objects[1].RADec.ToVector().Sub(c.Objects[0].RADec.ToVector())
	raDecAxis2 := Vector{-raDecAxis1.Y, raDecAxis1.X}
	x := v.Dot(imgAxis1)
	y := v.Dot(imgAxis2)
	if c.IsMirrored {
		y = -y
	}
	raDec := raDecAxis1.Times(x).Plus(raDecAxis2.Times(y))
	raDec = raDec.Plus(c.Objects[0].RADec.ToVector())
	return raDec.ToRADec()
}

func (c CoordinateSystem) RADecToImagePoint(raDec RADec) Vector {
	s := math.Abs(math.Cos(raDec.Dec / 180 * math.Pi))
	raDec.RA *= s
	c.Objects[0].RADec.RA *= s
	c.Objects[1].RADec.RA *= s
	v := raDec.ToVector().Sub(c.Objects[0].RADec.ToVector())
	raDecAxis1 := c.Objects[1].RADec.ToVector().Sub(c.Objects[0].RADec.ToVector())
	raDecAxis1 = raDecAxis1.Times(1 / raDecAxis1.Dot(raDecAxis1))
	raDecAxis2 := Vector{-raDecAxis1.Y, raDecAxis1.X}
	imgAxis1 := c.AlignmentPoints[1].Point.Sub(c.AlignmentPoints[0].Point)
	imgAxis2 := Vector{-imgAxis1.Y, imgAxis1.X}
	x := v.Dot(raDecAxis1)
	y := v.Dot(raDecAxis2)
	if c.IsMirrored {
		y = -y
	}
	imgVec := imgAxis1.Times(x).Plus(imgAxis2.Times(y))
	return c.AlignmentPoints[0].Point.Plus(imgVec)
}

var tap = "http://simbad.u-strasbg.fr/simbad/sim-tap/sync"

func tapQuery(query string) []byte {
	v := url.Values{}
	v.Add("request", "doQuery")
	v.Add("lang", "adql")
	v.Add("format", "json")
	v.Add("maxrec", "500")
	v.Add("query", query)
	resp, _ := http.Post(tap, "application/x-www-form-urlencoded", strings.NewReader(v.Encode()))
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

func generateTagsHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := ioutil.ReadAll(r.Body)
	data := struct {
		AspectRatio float64
		Points      [2]AlignmentPoint
		IsMirrored  bool
	}{}
	json.Unmarshal(body, &data)
	objects := [2]Object{}
	for i := range data.Points {
		body := tapQuery(`
			SELECT main_id, RA, DEC
			FROM basic JOIN ident ON oidref = oid
			WHERE id = '` + data.Points[i].ObjectId + `';`,
		)
		queryData := struct {
			Data [][]interface{}
		}{}
		json.Unmarshal(body, &queryData)
		objects[i] = Object{
			MainId: queryData.Data[0][0].(string),
			RADec: RADec{
				queryData.Data[0][1].(float64),
				queryData.Data[0][2].(float64),
			},
		}
	}
	c := CoordinateSystem{
		AspectRatio:     data.AspectRatio,
		AlignmentPoints: data.Points,
		Objects:         objects,
		IsMirrored:      data.IsMirrored,
	}
	origin := c.ImagePointToRADec(Vector{0, 0}).ToVector()
	//raDecVec := c.ImagePointToRADec(Vector{1, 1 / c.AspectRatio}).ToVector().Sub(origin)
	//center := raDecVec.Times(0.5).Plus(origin).ToRADec()
	topLeft := origin.ToRADec()
	topRight := c.ImagePointToRADec(Vector{1, 0})
	bottomRight := c.ImagePointToRADec(Vector{1, 1 / c.AspectRatio})
	bottomLeft := c.ImagePointToRADec(Vector{0, 1 / c.AspectRatio})
	query := fmt.Sprintf(`
		SELECT main_id, flux, RA, DEC
		FROM basic
		JOIN flux ON flux.oidref = oid
		WHERE filter = 'B' AND
			  flux < 12.0 AND
			  CONTAINS(POINT('ICRS', RA, DEC), POLYGON('ICRS',
			      %f, %f,
			      %f, %f,
			      %f, %f,
			      %f, %f
			  )) = 1
		ORDER BY flux ASC;`,
		topLeft.RA, topLeft.Dec,
		topRight.RA, topRight.Dec,
		bottomRight.RA, bottomRight.Dec,
		bottomLeft.RA, bottomLeft.Dec,
	)
	body = tapQuery(query)
	queryData := struct {
		Data [][]interface{}
	}{}
	json.Unmarshal(body, &queryData)
	query = fmt.Sprintf(`
		SELECT id, RA, DEC, galdim_majaxis, galdim_minaxis
		FROM basic
		JOIN ident ON ident.oidref = oid
		WHERE CONTAINS(POINT('ICRS', RA, DEC), POLYGON('ICRS',
			      %f, %f,
			      %f, %f,
			      %f, %f,
			      %f, %f
			  )) = 1 AND
			  id LIKE 'M %%';`,
		topLeft.RA, topLeft.Dec,
		topRight.RA, topRight.Dec,
		bottomRight.RA, bottomRight.Dec,
		bottomLeft.RA, bottomLeft.Dec,
	)
	body = tapQuery(query)
	mData := struct {
		Data [][]interface{}
	}{}
	json.Unmarshal(body, &mData)
	tags := []Tag{}
	for i := range mData.Data {
		obj := Object{
			MainId:    mData.Data[i][0].(string),
			Magnitude: 0,
			RADec: RADec{
				mData.Data[i][1].(float64),
				mData.Data[i][2].(float64),
			},
		}
		/*
			galDim := (mData.Data[i][3].(float64) + mData.Data[i][4].(float64)) / 2 / 60 / 60
			if obj.MainId == "M  45" {
				galDim = 80.0 / 60
			} else if obj.MainId == "M  31" {
				galDim = (178.0 + 63.0) / 2 / 60
			}
			dim := c.RADecToImagePoint(RADec{galDim + origin.X, origin.Y})
		*/
		tags = append(tags, Tag{
			Name:      obj.MainId,
			Magnitude: obj.Magnitude,
			Point:     c.RADecToImagePoint(obj.RADec),
			Dim:       0,
		})
	}
	for i := range queryData.Data {
		obj := Object{
			MainId:    queryData.Data[i][0].(string),
			Magnitude: queryData.Data[i][1].(float64),
			RADec: RADec{
				queryData.Data[i][2].(float64),
				queryData.Data[i][3].(float64),
			},
		}
		tags = append(tags, Tag{
			Name:      obj.MainId,
			Magnitude: obj.Magnitude,
			Point:     c.RADecToImagePoint(obj.RADec),
			Dim:       0,
		})
	}
	log.Print(len(tags))
	j, _ := json.Marshal(tags)
	w.Write(j)
}
