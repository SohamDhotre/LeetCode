class Solution {
    public int findCenter(int[][] edges) {
        int node1=edges[0][0], node2=edges[0][1];
        return (edges[1][0]==node1 || edges[1][1]==node1)?node1:node2;
    }
}