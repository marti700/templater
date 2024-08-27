from docx import Document
from docx.enum.text import WD_ALIGN_PARAGRAPH

# document = Document("EjemploGenerales.docx")
# paragraph = document.paragraphs[0]
# print(paragraph.text)

# document = Document("EjemploGenerales.docx")
# with open('resume.xml', 'w') as f:
# 	f.write(document._element.xml)


document = Document()
document.add_heading('Document Title', 0)

p = document.add_paragraph('Entre de una parte El senor ')
p.add_run('Teodoro Alcantara Aquino ').bold = True
p.add_run('Dominicano, mayor de edad, soltero, empleado privado, portador de la cedula de identidad y electoral No. 40221849o15, domiciliado y recidente en Santo Domingo;')
p.add_run('quien en lo que sigue del presente contrato se denominara ')
p.add_run('EL VENDEDOR').bold=True
p.alignment = WD_ALIGN_PARAGRAPH.JUSTIFY
document.save('demo.docx')