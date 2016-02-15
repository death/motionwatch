package detector

import "github.com/lazywei/go-opencv/opencv"

type Params struct {
	DevThreshold   float64
	PhaseThreshold int
	QueryFrame     func() *opencv.IplImage
}

type Detector struct {
	params Params

	previousFrame *opencv.IplImage
	currentFrame  *opencv.IplImage
	nextFrame     *opencv.IplImage
	d1            *opencv.IplImage
	d2            *opencv.IplImage
	motion        *opencv.IplImage

	phase int
}

func (d *Detector) Detect() *opencv.IplImage {
	opencv.Copy(d.currentFrame, d.previousFrame, nil)
	opencv.Copy(d.nextFrame, d.currentFrame, nil)
	opencv.CvtColor(d.params.QueryFrame(), d.nextFrame, opencv.CV_BGR2GRAY)

	opencv.AbsDiff(d.previousFrame, d.nextFrame, d.d1)
	opencv.AbsDiff(d.nextFrame, d.currentFrame, d.d2)
	opencv.BitwiseAnd(d.d1, d.d2, d.motion, nil)
	opencv.Threshold(d.motion, d.motion, 35, 255, opencv.CV_THRESH_BINARY)

	_, sdev := opencv.MeanStdDev(d.motion, nil)
	if sdev.Val()[0] <= d.params.DevThreshold {
		d.phase = 0
		return nil
	}

	d.phase++
	if d.phase < d.params.PhaseThreshold {
		return nil
	}

	d.phase = 0
	return d.params.QueryFrame()
}

func (d *Detector) Close() {
	d.motion.Release()
	d.d2.Release()
	d.d1.Release()
	d.nextFrame.Release()
	d.currentFrame.Release()
	d.previousFrame.Release()
}

func New(params *Params) *Detector {
	previousFrame := cloneGreyscale(params.QueryFrame())
	currentFrame := cloneGreyscale(params.QueryFrame())
	nextFrame := cloneGreyscale(params.QueryFrame())

	d1 := nextFrame.Clone()
	d2 := nextFrame.Clone()
	motion := nextFrame.Clone()

	return &Detector{
		params:        *params,
		previousFrame: previousFrame,
		currentFrame:  currentFrame,
		nextFrame:     nextFrame,
		d1:            d1,
		d2:            d2,
		motion:        motion,
	}
}

func cloneGreyscale(im *opencv.IplImage) *opencv.IplImage {
	w, h := im.Width(), im.Height()
	g := opencv.CreateImage(w, h, opencv.IPL_DEPTH_8U, 1)
	opencv.CvtColor(im, g, opencv.CV_BGR2GRAY)
	return g
}
