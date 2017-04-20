width, height = 500, 500

flag = "flag{heres_a_ppm}"

ppm = "P3 %s %s 255\n" % (width, height)

i = 0
for x in range(width):
    for y in range(height):
        lsb = flag[i % len(flag)]
        ppm += "%s %s %s\n" % (x%255, y%255, ord(lsb))
        i += 1

f = open("out.ppm", "w")
f.write(ppm)
f.close()
