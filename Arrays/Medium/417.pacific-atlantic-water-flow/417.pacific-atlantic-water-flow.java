class Solution {
    boolean [][]atlantic;
    boolean [][]pacific;
    int [][]heights;
    int m,n;
    public List<List<Integer>> pacificAtlantic(int[][] heights) {
        this.heights=heights;
        m=heights.length;
        n=heights[0].length;
        atlantic=new boolean[m][n];
        pacific=new boolean[m][n];
        List<List<Integer>>list=new ArrayList<>();

        for(int col=0;col<n;col++) dfs(pacific, 0, col);
        for(int row=0;row<m;row++) dfs(pacific, row, 0);

        for(int col=0;col<n;col++) dfs(atlantic, m-1, col);
        for(int row=0;row<m;row++) dfs(atlantic, row, n-1);
        
        for(int row=0;row<m;row++){
            for(int col=0;col<n;col++){
                if(atlantic[row][col] && pacific[row][col]){
                    list.add(List.of(row, col));
                }
            }
        }

        return list;
    }

    void dfs(boolean [][]ocean, int row, int col){
        if(ocean[row][col]) return;
        ocean[row][col]=true;

        int [][]dir={{-1, 0}, {0, -1}, {0, 1}, {1, 0}};
        for(int []d:dir){
            int nr=d[0]+row;
            int nc=d[1]+col;
            if(nr<0 || nc<0 || nr>=m || nc>=n) continue;
            if(heights[nr][nc]>=heights[row][col]) dfs(ocean, nr, nc);
        }
    }
}