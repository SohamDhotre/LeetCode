class Solution {
    public int orangesRotting(int[][] grid) {
        Deque<int[]>q=new ArrayDeque<>();
        int m=grid.length, n=grid[0].length, min=0, freshCount=0;
        int [][]dir={{-1,0},{0,1},{1,0},{0,-1}};
        for(int i=0;i<m;i++){
            for(int j=0;j<n;j++){
                if(grid[i][j]==1) freshCount++;
                else if(grid[i][j]==2) q.offer(new int[]{i,j});
            }
        }
        if (freshCount == 0) return 0;

        while(!q.isEmpty() && freshCount>0){
            int size=q.size();
            for(int itr=0;itr<size;itr++){
                int []cell=q.poll();
                int r=cell[0], c=cell[1];
                for(int i=0;i<dir.length;i++){
                    int nr=r+dir[i][0], nc=c+dir[i][1];
                    if(nr<0 || nr>=m || nc<0 || nc>=n) continue;
                    else if(grid[nr][nc]==1){
                        grid[nr][nc] = 2;
                        freshCount--;  
                        q.offer(new int[]{nr, nc});
                    }
                }
            }
            min++;
        }
        return freshCount==0?min:-1;
    }
}