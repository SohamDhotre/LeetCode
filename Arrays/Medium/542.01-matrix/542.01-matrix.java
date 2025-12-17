class Solution {
    public int[][] updateMatrix(int[][] mat) {
        int m=mat.length, n=mat[0].length;
        int [][]dist=new int[m][n];
        Queue<int[]>q=new ArrayDeque<>();

        for(int i=0;i<m;i++){
            for(int j=0;j<n;j++){
                if(mat[i][j]==0){
                    dist[i][j]=0;
                    q.offer(new int[]{i,j});
                }
                else dist[i][j]=Integer.MAX_VALUE;
            }
        }

        int [][]dirs=new int[][]{
            {1,0},
            {-1,0},
            {0,1},
            {0,-1}
        };
        
        while(!q.isEmpty()){
            int []cur=q.poll();
            int r=cur[0], c=cur[1];

            for(int []dir:dirs){
                int nr=r+dir[0], nc=c+dir[1];
                if(nr<0 || nc<0 || nr>=m || nc>=n) continue;
                if(dist[nr][nc]>dist[r][c]+1){
                    dist[nr][nc]=dist[r][c]+1;
                    q.offer(new int[]{nr, nc});
                }
            }
        }

        return dist;
    }
}