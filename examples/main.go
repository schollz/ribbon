package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"time"

	. "github.com/fogleman/fauxgl"
	"github.com/fogleman/ribbon"
	"github.com/nfnt/resize"
)

const (
	scale  = 4
	width  = 1600 * 1
	height = 1600 * 1
	fovy   = 25
	near   = 1
	far    = 10
)

// var (
// 	eye    = V(4, 0, 4)
// 	center = V(0, -0.03, 0)
// 	up     = V(0, 1, 0).Normalize()
// 	light  = V(0.75, 0.25, 0.25).Normalize()
// )

var (
	eye    = V(5, 0.5, 0)
	center = V(0, -0.025, 0)
	up     = V(0, 1, 0).Normalize()
	light  = V(0.75, 0.25, 0.25).Normalize()
)

func makeCylinder(p0, p1 Vector, r float64) *Mesh {
	p := p0.Add(p1).MulScalar(0.5)
	h := p0.Distance(p1) * 2
	up := p1.Sub(p0).Normalize()
	mesh := NewCylinder(15, false)
	mesh.Transform(Orient(p, V(r, r, h), up, 0))
	return mesh
}

func dumpMesh(mesh *Mesh) {
	var vertices []Vector
	lookup := make(map[Vector]int)
	// colors := make(map[Vector]Color)
	var colors []Color
	colorLookup := make(map[Color]int)
	for _, t := range mesh.Triangles {
		if _, ok := lookup[t.V1.Position]; !ok {
			lookup[t.V1.Position] = len(vertices)
			vertices = append(vertices, t.V1.Position)
		}
		if _, ok := lookup[t.V2.Position]; !ok {
			lookup[t.V2.Position] = len(vertices)
			vertices = append(vertices, t.V2.Position)
		}
		if _, ok := lookup[t.V3.Position]; !ok {
			lookup[t.V3.Position] = len(vertices)
			vertices = append(vertices, t.V3.Position)
		}
		if _, ok := colorLookup[t.V1.Color]; !ok {
			colorLookup[t.V1.Color] = len(colors)
			colors = append(colors, t.V1.Color)
		}
		if _, ok := colorLookup[t.V2.Color]; !ok {
			colorLookup[t.V2.Color] = len(colors)
			colors = append(colors, t.V2.Color)
		}
		if _, ok := colorLookup[t.V3.Color]; !ok {
			colorLookup[t.V3.Color] = len(colors)
			colors = append(colors, t.V3.Color)
		}
	}
	fmt.Println("var VERTICES = [")
	for _, v := range vertices {
		fmt.Printf("[%.8f,%.8f,%.8f],\n", v.X, v.Y, v.Z)
	}
	fmt.Println("];")
	fmt.Println("var COLORS = [")
	for _, c := range colors {
		fmt.Printf("[%.3f,%.3f,%.3f],\n", c.R, c.G, c.B)
	}
	fmt.Println("];")
	fmt.Println("var FACES = [")
	for _, t := range mesh.Triangles {
		i1 := lookup[t.V1.Position]
		i2 := lookup[t.V2.Position]
		i3 := lookup[t.V3.Position]
		c1 := colorLookup[t.V1.Color]
		c2 := colorLookup[t.V2.Color]
		c3 := colorLookup[t.V3.Color]
		fmt.Printf("[%d,%d,%d,%d,%d,%d],\n", i1, i2, i3, c1, c2, c3)
	}
	fmt.Println("];")
}

func main() {
	model, err := ribbon.LoadPDB(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println(len(model.Atoms), len(model.Residues), len(model.Chains))

	mesh := NewEmptyMesh()
	for _, c := range model.Chains {
		m := c.Mesh()
		for i, t := range m.Triangles {
			p := float64(i) / float64(len(m.Triangles)-1)
			c := MakeColor(Viridis.Color(p))
			t.V1.Color = c
			t.V2.Color = c
			t.V3.Color = c
		}
		mesh.Add(m)
	}
	// fmt.Println(len(mesh.Triangles))

	sphere := NewSphere(15, 15)
	sphere.SmoothNormals()
	atomsBySerial := make(map[int]*ribbon.Atom)
	for _, a := range model.HetAtoms {
		if a.ResName == "HOH" {
			continue
		}
		atomsBySerial[a.Serial] = a
		e := a.GetElement()
		c := HexColor(e.HexColor)
		r := e.Radius * 0.75
		s := V(r, r, r)
		m := sphere.Copy()
		m.Transform(Scale(s).Translate(a.Position))
		for _, t := range m.Triangles {
			t.V1.Color = c
			t.V2.Color = c
			t.V3.Color = c
		}
		mesh.Add(m)
	}
	// fmt.Println(len(mesh.Triangles))

	for _, c := range model.Connections {
		a1 := atomsBySerial[c.Serial1]
		a2 := atomsBySerial[c.Serial2]
		if a1 == nil || a2 == nil {
			continue
		}
		e1 := a1.GetElement()
		e2 := a2.GetElement()
		p1 := a1.Position.LerpDistance(a2.Position, e1.Radius*0.75-0.1)
		p2 := a2.Position.LerpDistance(a1.Position, e2.Radius*0.75-0.1)
		mid := p1.Lerp(p2, 0.5)
		m := makeCylinder(p1, mid, 0.25)
		c := HexColor(e1.HexColor)
		for _, t := range m.Triangles {
			t.V1.Color = c
			t.V2.Color = c
			t.V3.Color = c
		}
		mesh.Add(m)
		m = makeCylinder(p2, mid, 0.25)
		c = HexColor(e2.HexColor)
		for _, t := range m.Triangles {
			t.V1.Color = c
			t.V2.Color = c
			t.V3.Color = c
		}
		mesh.Add(m)
	}
	// fmt.Println(len(mesh.Triangles))

	// base := mesh.Copy()
	// for _, matrix := range model.SymmetryMatrixes {
	// 	if matrix == Identity() {
	// 		continue
	// 	}
	// 	m := base.Copy()
	// 	m.Transform(matrix)
	// 	mesh.Add(m)
	// }
	// fmt.Println(len(mesh.Triangles))

	mesh.BiUnitCube()
	dumpMesh(mesh)
	return

	// mesh.SmoothNormalsThreshold(Radians(75))
	// mesh.SaveSTL("out.stl")

	// create a rendering context
	context := NewContext(width*scale, height*scale)
	context.ClearColorBufferWith(HexColor("1D181F"))

	// create transformation matrix and light direction
	aspect := float64(width) / float64(height)
	matrix := LookAt(eye, center, up).Perspective(fovy, aspect, near, far)

	// render
	shader := NewPhongShader(matrix, light, eye)
	shader.AmbientColor = Gray(0.3)
	shader.DiffuseColor = Gray(0.9)
	context.Shader = shader
	// context.Cull = CullFront
	start := time.Now()
	context.DrawTriangles(mesh.Triangles)
	fmt.Println(time.Since(start))

	// context.ClearDepthBuffer()
	// start = time.Now()
	// context.Cull = CullBack
	// context.DrawTriangles(mesh.Triangles)
	// fmt.Println(time.Since(start))

	// save image
	image := context.Image()
	image = resize.Resize(width, height, image, resize.Bilinear)
	SavePNG("out.png", image)

	// for i := 0; i < 720; i += 1 {
	// 	context.ClearColorBufferWith(HexColor("1D181F"))
	// 	context.ClearDepthBuffer()

	// 	shader := NewPhongShader(matrix, light, eye)
	// 	shader.AmbientColor = Gray(0.3)
	// 	shader.DiffuseColor = Gray(0.9)
	// 	context.Shader = shader
	// 	start := time.Now()
	// 	context.DepthBias = 0
	// 	context.DrawTriangles(mesh.Triangles)
	// 	fmt.Println(time.Since(start))

	// 	image := context.Image()
	// 	image = resize.Resize(width, height, image, resize.Bilinear)
	// 	SavePNG(fmt.Sprintf("frame%03d.png", i), image)

	// 	mesh.Transform(Rotate(up, Radians(0.5)))
	// }
}

type Colormap struct {
	Colors []color.Color
}

func NewColormap(colors []color.Color) *Colormap {
	return &Colormap{colors}
}

func (c Colormap) Color(t float64) color.Color {
	n := len(c.Colors)
	i := int(math.Floor(t * float64(n)))
	if i < 0 {
		i = 0
	}
	if i > n-1 {
		i = n - 1
	}
	return c.Colors[i]
}

func parseColors(x string) []color.Color {
	var result []color.Color
	for i := 0; i < len(x); i += 6 {
		var r, g, b int
		fmt.Sscanf(x[i:i+6], "%02x%02x%02x", &r, &g, &b)
		c := color.NRGBA{uint8(r), uint8(g), uint8(b), 255}
		result = append(result, c)
	}
	return result
}

var Viridis = NewColormap(parseColors("44015444025645045745055946075a46085c460a5d460b5e470d60470e6147106347116447136548146748166848176948186a481a6c481b6d481c6e481d6f481f70482071482173482374482475482576482677482878482979472a7a472c7a472d7b472e7c472f7d46307e46327e46337f463480453581453781453882443983443a83443b84433d84433e85423f854240864241864142874144874045884046883f47883f48893e49893e4a893e4c8a3d4d8a3d4e8a3c4f8a3c508b3b518b3b528b3a538b3a548c39558c39568c38588c38598c375a8c375b8d365c8d365d8d355e8d355f8d34608d34618d33628d33638d32648e32658e31668e31678e31688e30698e306a8e2f6b8e2f6c8e2e6d8e2e6e8e2e6f8e2d708e2d718e2c718e2c728e2c738e2b748e2b758e2a768e2a778e2a788e29798e297a8e297b8e287c8e287d8e277e8e277f8e27808e26818e26828e26828e25838e25848e25858e24868e24878e23888e23898e238a8d228b8d228c8d228d8d218e8d218f8d21908d21918c20928c20928c20938c1f948c1f958b1f968b1f978b1f988b1f998a1f9a8a1e9b8a1e9c891e9d891f9e891f9f881fa0881fa1881fa1871fa28720a38620a48621a58521a68522a78522a88423a98324aa8325ab8225ac8226ad8127ad8128ae8029af7f2ab07f2cb17e2db27d2eb37c2fb47c31b57b32b67a34b67935b77937b87838b9773aba763bbb753dbc743fbc7340bd7242be7144bf7046c06f48c16e4ac16d4cc26c4ec36b50c46a52c56954c56856c66758c7655ac8645cc8635ec96260ca6063cb5f65cb5e67cc5c69cd5b6ccd5a6ece5870cf5773d05675d05477d1537ad1517cd2507fd34e81d34d84d44b86d54989d5488bd6468ed64590d74393d74195d84098d83e9bd93c9dd93ba0da39a2da37a5db36a8db34aadc32addc30b0dd2fb2dd2db5de2bb8de29bade28bddf26c0df25c2df23c5e021c8e020cae11fcde11dd0e11cd2e21bd5e21ad8e219dae319dde318dfe318e2e418e5e419e7e419eae51aece51befe51cf1e51df4e61ef6e620f8e621fbe723fde725"))
var Magma = NewColormap(parseColors("00000401000501010601010802010902020b02020d03030f03031204041405041606051806051a07061c08071e0907200a08220b09240c09260d0a290e0b2b100b2d110c2f120d31130d34140e36150e38160f3b180f3d19103f1a10421c10441d11471e114920114b21114e22115024125325125527125829115a2a115c2c115f2d11612f116331116533106734106936106b38106c390f6e3b0f703d0f713f0f72400f74420f75440f764510774710784910784a10794c117a4e117b4f127b51127c52137c54137d56147d57157e59157e5a167e5c167f5d177f5f187f601880621980641a80651a80671b80681c816a1c816b1d816d1d816e1e81701f81721f817320817521817621817822817922827b23827c23827e24828025828125818326818426818627818827818928818b29818c29818e2a81902a81912b81932b80942c80962c80982d80992d809b2e7f9c2e7f9e2f7fa02f7fa1307ea3307ea5317ea6317da8327daa337dab337cad347cae347bb0357bb2357bb3367ab5367ab73779b83779ba3878bc3978bd3977bf3a77c03a76c23b75c43c75c53c74c73d73c83e73ca3e72cc3f71cd4071cf4070d0416fd2426fd3436ed5446dd6456cd8456cd9466bdb476adc4869de4968df4a68e04c67e24d66e34e65e44f64e55064e75263e85362e95462ea5661eb5760ec5860ed5a5fee5b5eef5d5ef05f5ef1605df2625df2645cf3655cf4675cf4695cf56b5cf66c5cf66e5cf7705cf7725cf8745cf8765cf9785df9795df97b5dfa7d5efa7f5efa815ffb835ffb8560fb8761fc8961fc8a62fc8c63fc8e64fc9065fd9266fd9467fd9668fd9869fd9a6afd9b6bfe9d6cfe9f6dfea16efea36ffea571fea772fea973feaa74feac76feae77feb078feb27afeb47bfeb67cfeb77efeb97ffebb81febd82febf84fec185fec287fec488fec68afec88cfeca8dfecc8ffecd90fecf92fed194fed395fed597fed799fed89afdda9cfddc9efddea0fde0a1fde2a3fde3a5fde5a7fde7a9fde9aafdebacfcecaefceeb0fcf0b2fcf2b4fcf4b6fcf6b8fcf7b9fcf9bbfcfbbdfcfdbf"))
var Inferno = NewColormap(parseColors("00000401000501010601010802010a02020c02020e03021004031204031405041706041907051b08051d09061f0a07220b07240c08260d08290e092b10092d110a30120a32140b34150b37160b39180c3c190c3e1b0c411c0c431e0c451f0c48210c4a230c4c240c4f260c51280b53290b552b0b572d0b592f0a5b310a5c320a5e340a5f3609613809623909633b09643d09653e0966400a67420a68440a68450a69470b6a490b6a4a0c6b4c0c6b4d0d6c4f0d6c510e6c520e6d540f6d550f6d57106e59106e5a116e5c126e5d126e5f136e61136e62146e64156e65156e67166e69166e6a176e6c186e6d186e6f196e71196e721a6e741a6e751b6e771c6d781c6d7a1d6d7c1d6d7d1e6d7f1e6c801f6c82206c84206b85216b87216b88226a8a226a8c23698d23698f24699025689225689326679526679727669827669a28659b29649d29649f2a63a02a63a22b62a32c61a52c60a62d60a82e5fa92e5eab2f5ead305dae305cb0315bb1325ab3325ab43359b63458b73557b93556ba3655bc3754bd3853bf3952c03a51c13a50c33b4fc43c4ec63d4dc73e4cc83f4bca404acb4149cc4248ce4347cf4446d04545d24644d34743d44842d54a41d74b3fd84c3ed94d3dda4e3cdb503bdd513ade5238df5337e05536e15635e25734e35933e45a31e55c30e65d2fe75e2ee8602de9612bea632aeb6429eb6628ec6726ed6925ee6a24ef6c23ef6e21f06f20f1711ff1731df2741cf3761bf37819f47918f57b17f57d15f67e14f68013f78212f78410f8850ff8870ef8890cf98b0bf98c0af98e09fa9008fa9207fa9407fb9606fb9706fb9906fb9b06fb9d07fc9f07fca108fca309fca50afca60cfca80dfcaa0ffcac11fcae12fcb014fcb216fcb418fbb61afbb81dfbba1ffbbc21fbbe23fac026fac228fac42afac62df9c72ff9c932f9cb35f8cd37f8cf3af7d13df7d340f6d543f6d746f5d949f5db4cf4dd4ff4df53f4e156f3e35af3e55df2e661f2e865f2ea69f1ec6df1ed71f1ef75f1f179f2f27df2f482f3f586f3f68af4f88ef5f992f6fa96f8fb9af9fc9dfafda1fcffa4"))
var Plasma = NewColormap(parseColors("0d088710078813078916078a19068c1b068d1d068e20068f2206902406912605912805922a05932c05942e05952f059631059733059735049837049938049a3a049a3c049b3e049c3f049c41049d43039e44039e46039f48039f4903a04b03a14c02a14e02a25002a25102a35302a35502a45601a45801a45901a55b01a55c01a65e01a66001a66100a76300a76400a76600a76700a86900a86a00a86c00a86e00a86f00a87100a87201a87401a87501a87701a87801a87a02a87b02a87d03a87e03a88004a88104a78305a78405a78606a68707a68808a68a09a58b0aa58d0ba58e0ca48f0da4910ea3920fa39410a29511a19613a19814a099159f9a169f9c179e9d189d9e199da01a9ca11b9ba21d9aa31e9aa51f99a62098a72197a82296aa2395ab2494ac2694ad2793ae2892b02991b12a90b22b8fb32c8eb42e8db52f8cb6308bb7318ab83289ba3388bb3488bc3587bd3786be3885bf3984c03a83c13b82c23c81c33d80c43e7fc5407ec6417dc7427cc8437bc9447aca457acb4679cc4778cc4977cd4a76ce4b75cf4c74d04d73d14e72d24f71d35171d45270d5536fd5546ed6556dd7566cd8576bd9586ada5a6ada5b69db5c68dc5d67dd5e66de5f65de6164df6263e06363e16462e26561e26660e3685fe4695ee56a5de56b5de66c5ce76e5be76f5ae87059e97158e97257ea7457eb7556eb7655ec7754ed7953ed7a52ee7b51ef7c51ef7e50f07f4ff0804ef1814df1834cf2844bf3854bf3874af48849f48948f58b47f58c46f68d45f68f44f79044f79143f79342f89441f89540f9973ff9983ef99a3efa9b3dfa9c3cfa9e3bfb9f3afba139fba238fca338fca537fca636fca835fca934fdab33fdac33fdae32fdaf31fdb130fdb22ffdb42ffdb52efeb72dfeb82cfeba2cfebb2bfebd2afebe2afec029fdc229fdc328fdc527fdc627fdc827fdca26fdcb26fccd25fcce25fcd025fcd225fbd324fbd524fbd724fad824fada24f9dc24f9dd25f8df25f8e125f7e225f7e425f6e626f6e826f5e926f5eb27f4ed27f3ee27f3f027f2f227f1f426f1f525f0f724f0f921"))
