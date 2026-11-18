# Desensambla palabras "ADDR WORD" (hex) de un fichero: dis2.py fichero desde hasta
import sys
R=['zero','at','v0','v1','a0','a1','a2','a3','t0','t1','t2','t3','t4','t5','t6','t7','s0','s1','s2','s3','s4','s5','s6','s7','t8','t9','k0','k1','gp','sp','fp','ra']
def dis(a,i):
    op=i>>26; rs=(i>>21)&31; rt=(i>>16)&31; rd=(i>>11)&31; sh=(i>>6)&31; fn=i&63; imm=i&0xffff; simm=imm-0x10000 if imm&0x8000 else imm
    if i==0: return 'nop'
    if op==0:
        names={0:'sll',2:'srl',3:'sra',4:'sllv',6:'srlv',7:'srav',8:'jr',9:'jalr',16:'mfhi',18:'mflo',24:'mult',25:'multu',26:'div',27:'divu',32:'add',33:'addu',34:'sub',35:'subu',36:'and',37:'or',38:'xor',39:'nor',42:'slt',43:'sltu'}
        n=names.get(fn,'sp%d'%fn)
        if fn in (0,2,3): return '%s %s,%s,%d'%(n,R[rd],R[rt],sh)
        if fn==8: return 'jr %s'%R[rs]
        if fn==9: return 'jalr %s'%R[rs]
        if fn in (16,18): return '%s %s'%(n,R[rd])
        if fn in (24,25,26,27): return '%s %s,%s'%(n,R[rs],R[rt])
        return '%s %s,%s,%s'%(n,R[rd],R[rs],R[rt])
    names={2:'j',3:'jal',4:'beq',5:'bne',6:'blez',7:'bgtz',8:'addi',9:'addiu',10:'slti',11:'sltiu',12:'andi',13:'ori',14:'xori',15:'lui',32:'lb',33:'lh',35:'lw',36:'lbu',37:'lhu',40:'sb',41:'sh',43:'sw',18:'cop2',50:'lwc2',58:'swc2'}
    n=names.get(op,'op%d'%op)
    if op in (2,3): return '%s 0x%08x'%(n,(a&0xf0000000)|((i&0x3ffffff)<<2))
    if op in (4,5): return '%s %s,%s,0x%08x'%(n,R[rs],R[rt],a+4+simm*4)
    if op in (6,7): return '%s %s,0x%08x'%(n,R[rs],a+4+simm*4)
    if op==1: return '%s %s,0x%08x'%({0:'bltz',1:'bgez',16:'bltzal',17:'bgezal'}.get(rt,'b?'),R[rs],a+4+simm*4)
    if op>=32: return '%s %s,%d(%s)'%(n,R[rt],simm,R[rs])
    if op==15: return 'lui %s,0x%x'%(R[rt],imm)
    if op in (12,13,14): return '%s %s,%s,0x%x'%(n,R[rt],R[rs],imm)
    return '%s %s,%s,%d'%(n,R[rt],R[rs],simm)
lo=int(sys.argv[2],16); hi=int(sys.argv[3],16)
for l in open(sys.argv[1]):
    a,i=l.split(); a=int(a,16); i=int(i,16)
    if lo<=a<=hi: print('%08x: %08x  %s'%(a,i,dis(a,i)))
