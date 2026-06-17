from PIL import Image
import math

def distance(c1, c2):
    return math.sqrt(sum((a - b) ** 2 for a, b in zip(c1, c2)))

img = Image.open('frontend/public/logo-banner.png').convert("RGBA")
datas = img.getdata()

target_color = (229, 230, 233)
newData = []

for item in datas:
    if distance(item[:3], target_color) < 20: # Tolerance
        newData.append((255, 255, 255, 0))
    else:
        newData.append(item)

img.putdata(newData)
img.save('frontend/public/logo-banner.png', "PNG")
print("Logo background made transparent.")
