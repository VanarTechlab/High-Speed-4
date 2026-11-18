import struct,sys
d=open('disc/SLUS_008.26','rb').read()
t_addr=struct.unpack_from('<I',d,0x18)[0]
R=['zero','at','v0','v1','a0','a1','a2','a3','t0','t1','t2','t3','t4','t5','t6','t7','s0','s1','s2','s3','s4','s5','s6','s7','t8','t9','k0','k1','gp','sp','fp','ra']
def w(a): return struct.unpack_from('<I',d,0x800+a-t_addr)[0]
def dis(a,i):
    op=i>>26; rs=(i>>21)&31; rt=(i>>16)&31; rd=(i>>11)&31; sh=(i>>6)&31; fn=i&63; imm=i&0xffff; simm=imm-0x10000 if imm&0x8000 else imm
    if i==0: return 'nop'
    if op==0:
        names={0:'sll',2:'srl',3:'sra',4:'sllv',6:'srlv',7:'srav',8:'jr',9:'jalr',16:'mfhi',18:'mflo',24:'mult',25:'multu',26:'div',27:'divu',32:'add',33:'addu',34:'sub',35:'subu',36:'and',37:'or',38:'xor',39:'nor',42:'slt',43:'sltu'}
        n=names.get(fn,'sp%d'%fn)
        if fn in (0,2,3): return '%s %s,%s,%d'%(n,R[rd],R[rt],sh)
        if fn==8: return 'jr %s'%R[rs]
        return '%s %s,%s,%s'%(n,R[rd],R[rs],R[rt])
    names={1:'bcond',2:'j',3:'jal',4:'beq',5:'bne',6:'blez',7:'bgtz',8:'addi',9:'addiu',10:'slti',11:'sltiu',12:'andi',13:'ori',15:'lui',32:'lb',33:'lh',35:'lw',36:'lbu',37:'lhu',40:'sb',41:'sh',43:'sw',18:'cop2',50:'lwc2',58:'swc2'}
    n=names.get(op,'op%d'%op)
    if op in (2,3): return '%s 0x%08x'%(n,(a&0xf0000000)|((i&0x3ffffff)<<2))
    if op in (4,5): return '%s %s,%s,0x%08x'%(n,R[rs],R[rt],a+4+simm*4)
    if op in (6,7): return '%s %s,0x%08x'%(n,R[rs],a+4+simm*4)
    if op==1: return '%s %s,%d,0x%08x'%('bltz' if rt==0 else 'bgez' if rt==1 else 'bcond%d'%rt,R[rs],rt,a+4+simm*4)
    if op>=32: return '%s %s,%d(%s)'%(n,R[rt],simm,R[rs])
    if op==15: return 'lui %s,0x%x'%(R[rt],imm)
    return '%s %s,%s,%d'%(n,R[rt],R[rs],simm)
a0=int(sys.argv[1],16); n=int(sys.argv[2])
for a in range(a0,a0+4*n,4):
    i=w(a); print('%08x: %08x  %s'%(a,i,dis(a,i)))
